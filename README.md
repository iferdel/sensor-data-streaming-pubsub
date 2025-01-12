# Sensor Data Streaming PubSub

## General Description
![grafana-dashboard](./assets/grafana-dashboard.gif)
* (GIF showing grafana-dashboard with more than one sensor + monitor db queries --stats from postgres using timescaledb functionality--)
* (GIF showing iotctl behaviour -- maybe with bubbletea implemented already which would beautify the status of running sensors and not running sensors)*
* (GIF showing pods on k8 -- invitation to [homelab](https://github.com/iferdel/homelab))
* *maybe(GIF showing map with GPS data from sensors -- either static or dynamic locations)*

> [!NOTE]
> I think that a design like this is a good starting point for a larger project that would involve a real and broader sensor monitoring spectrum with GPS (either via Wi-Fi or GSM), in mobile vehicles or static machinery aswell as in civil infrastructure.

## Reason
Back in 2020, I worked on **vibration analysis**. My main background at that time was in **Mechanical Engineering**, and I took on a role that involved designing sensor installations, performing in-field measurements, and analyzing the data back at the office. 

I measured various types of **mechanical equipment**, such as overhead cranes in a mining plant and climate control systems (including pumps, cooling towers and air handling units) at Chile’s main airport. Additionally, I assessed **civil structures**--protected, commercial, and private buildings--subjected to nearby construction or physical phenomena, such as vibrations generated by passing trains.

All of these tasks were performed **in-situ**, which motivated me to consider a more ambitious approach: **remote, real-time monitoring**. Such a system could open up **new business opportunities** by offering continuous insight without requiring on-site personnel.

With that in mind, my goal for this project is to build a comprehensive **end-to-end**, real-time monitoring solution.


## Architecture
*(architecture diagram)*

The core of this solution is based on an **event-driven** architecture using a **pub/sub** pattern at its core, making **distributed system** possible. Nevertheless, as with any other system, an **hybrid** approach is required, such as relying on **point-to-point** communication for the interaction with the sensor cluster thorugh a command line tooling *iotctl* which communicates with an *api* that enables a controlled interaction with the database and message broker.

### The services defined in the project are the following:
<dl>
  <dt><code>iotctl</code></dt>
  <dd>Command line tool to interact remotely with cluster of nodes.</dd>
  <dt><code>iot-api</code></dt>
  <dd>An API that facilitates communication between the service and *iotctl* users over *HTTPS*. It's like the gateway to interact with the database and the sensors themselves thourgh the message broker.</dd>
  <dt><code>sensor-simulation</code></dt>
  <dd>Simulates an accelerometer measuring on a specific environment, that means it mimics the signal of, for example, a bearing from a pump system. It consumes commands sent from iot-api and publishes its logs (like booting logs), the sensor serial number for registration of the sensor into the database aswell as the measurement values.</dd>
  <dt><code>sensor-registry</code></dt>
  <dd>It consumes the sensor information about serial number, like a 'look, I'm sensor with serial number xxxx, if I'm not in the database, go register me so i can start sending measurements".</dd>
  <dt><code>sensor-logs-ingester</code></dt>
  <dd>It consumes the sensor logs and saves them into a .log file to further processing in a centralized manner.</dd>
  <dt><code>sensor-measurements-ingester</code></dt>
  <dd>It consumes the sensor(s) measurements and insert them into the postgres/timescaledb instance.</dd>
</dl>

> [!IMPORTANT]
> These services are dependant of other software such as the message broker, a database that would handle timeseries data with ease and a visualization tool to real-time monitoring.

Having said that, the project is intended to be hosted on a **Kubernetes cluster** to ensure high availability and horizontal scaling as needed—for example, if the amount of incoming data increases from an increase in sensor sampling rates or if more sensors are added (which in the simulation sense, it is a kind of replication of the simulation service which could count as an horizontal scaling).
Yet, the only services that would not be hosted directly on the cluster is the database itself and the command line tool since it is intended to be installed on users that want to interact with the services.
Not fan of hosting databases in k8, I would definitely use the cloud solution as it scales better and there is a whole team taking care of this.

*Disclaimer: one could conclude that a hybrid architecture for critical low-latency control would also be quite handy. In that case, one would expect using gRPC as the way to communicate between a service that would send direct commands to change behaviour (in a reactive way) not the sensor but to the machine or whatever is behind.*

<details>
<summary><strong>:mag: Key Architectural Points</strong></summary>

- **Data Transfer**: The solution is intended to use Protobuf as a data serialization format to match real scenarios with embedded C or C++. However, for the initial setup (POC), the Go encoding/gob serializer is in use to ease development.
- **Infrastructure**: This project integrates with my [homelab](https://github.com/iferdel/homelab), which simulates a cloud-like environment on bare metal using TalosOS and GitOps with FluxCD. The only service that's out from the cluster is the command line tool which is intended to be used within a remote machine that needs to authenticate in order to interact with the sensor cluster.
- **CI/CD**: For CI/CD, I’m using a private Jenkins server and Docker Hub for image storage, while the GitHub repository hosts the source code. The whole CD would be handled with FluxCD.
- **Secrets**: I’m using Azure Key Vault for secrets in the homelab. 
- **Database**: The solution uses PostgreSQL with [TimeScaleDB](https://www.timescale.com/), an extension optimized for time-series data. In a real scenario, the paid cloud tier would be in use, but for this project I’m storaging everything on bare metal, integrated with CloudNativePG and OpenEBS + Mayastor for storage. Ephemeral and with data retention policy.
- **Data Management**: TimeScaleDB’s policies handle data expiration and compression, preventing storage overflow and improving performance.
- **Visualization**: Grafana is used for near real-time dashboards, leveraging its querying capabilities to visualize time-series data stored as well as stats from the database itself by means of wrapping the stats from pg_stat_statements and pg_stat_kcache with postgres CTEs and procedures.
- **Alarms**: *...*  
- **Communication Protocols**:
    - *Sensor communication uses MQTT with streaming queues.*
    - *Inter-service communication uses AMQP with RabbitMQ, employing quorum queues.*
    - *Alarm service communication uses gRPC for low-latency communication with the machine where the sensor to affect behaviour*

</details>

## :art: Design 

> Just threw some paint on the canvas tonight.
>
> -- <cite><i>Kawhi Leonard</i></cite>

Sensors will send:
    - id (serial number)
    - timestamp (with nanosecond precision)
    - measurement (at a variable sample frequency)
    - GPS location (latitude and longitude)
This approach allows flexible scaling of the number of sensors and their data rates.

Sensor will receive (mapped through its id):
    - commands that would affect the sensor behaviour such as sleep, awake and change sample frequency.

<details>
<summary><strong>:deciduous_tree: Directory Tree</strong></summary>

*I like the structure that became manifest while developing the project. That's why I'm attaching the filetree since it reads nicely.*
```
.
├── LICENSE
├── README.md
├── assets
│   └── grafana-dashboard.gif
├── cmd
│   ├── iotctl
│   │   ├── Dockerfile
│   │   ├── cmd
│   │   │   ├── awake.go
│   │   │   ├── changesamplefrequency.go
│   │   │   ├── delete.go
│   │   │   ├── root.go
│   │   │   ├── sensorstatus.go
│   │   │   └── sleep.go
│   │   └── main.go
│   ├── sensor-logs-ingester
│   │   ├── Dockerfile
│   │   ├── handlers.go
│   │   └── main.go
│   ├── sensor-measurements-ingester
│   │   ├── Dockerfile
│   │   ├── handlers.go
│   │   └── main.go
│   ├── sensor-registry
│   │   ├── Dockerfile
│   │   ├── handlers.go
│   │   └── main.go
│   └── sensor-simulation
│       ├── Dockerfile
│       ├── handlers.go
│       └── main.go
├── compose.yaml
├── dependencies
│   ├── grafana
│   │   ├── README.md
│   │   ├── grafana.ini
│   │   └── provisioning
│   │       ├── dashboards
│   │       │   ├── iot.json
│   │       │   ├── iot.yaml
│   │       │   └── queries.sql
│   │       └── datasources
│   │           └── datasources.yaml
│   ├── rabbitmq
│   │   ├── Dockerfile
│   │   ├── definitions.json
│   │   └── rabbitmq.conf
│   └── timescaledb
│       ├── Dockerfile
│       ├── init.sh
│       └── postgresql.conf
├── go.mod
├── go.sum
├── ideas.md
├── internal
│   ├── pubsub
│   │   ├── consume.go
│   │   └── publish.go
│   ├── routing
│   │   ├── models.go
│   │   └── routing.go
│   ├── sensorlogic
│   │   ├── awake.go
│   │   ├── changesamplefrequency.go
│   │   ├── sensor.go
│   │   ├── sensorlogs.go
│   │   ├── sensormeasurements.go
│   │   ├── sensorsignal.go
│   │   └── sleep.go
│   └── storage
│       ├── README.md
│       ├── db.go
│       ├── logs.go
│       ├── measurements.go
│       ├── models.go
│       └── sensors.go
└── utils
    └── wait-for-services.sh
```

</details>

<details open>
<summary><strong>:elephant: :tiger: Database Schema</strong></summary>

The beauty of TimescaleDB is that it’s built on top of PostgreSQL, allowing us to use SQL and thus embrace core principles of relational databases, such as normalization.

![database-erd](./assets/sensor-data-streaming-pubsub-erd.drawio.svg)


> [Timescale hypertables do not support primary keys](https://stackoverflow.com/a/77463051). This is because the underlying data must be partitioned to several physical PostgreSQL tables. Partitioned look-ups cannot support a primary key, but a [composite primary key](https://docs.timescale.com/use-timescale/latest/schema-management/about-constraints/#about-constraints) of together unique columns could be used.

</details>

<details open>
<summary><strong>:rabbit: Messaging Routing</strong></summary>

Exchange, Queues, and Routing Keys:

    Exchange of type Topic: iot
    Queues, following entity.id.consumer.type pattern:
        - sensor.all.measurements.db_writer
        - sensor.<sensor.serial_number>.commands               
        - sensor.all.registry.created      
        - sensor.all.logs
    Keys used in consumers with wildcards and in publishers with the specific value
        Publishers:
        - sensor.<sensor.serial_number>.measurements
        - sensor.<sensor.serial_number>.commands
        - sensor.<sensor.serial_number>.registry
        - sensor.<sensor.serial_number>.logs
        Consumers:
        - sensor.*.measurements
        - sensor.*.commands.#
        - sensor.*.registry.#
        - sensor.*.logs.#

</summary>

<details open>
<summary><strong>:pencil: Engineering Calculation Report</strong></summary>

**General Formula of Accelerometer Signal**\
$`a(t) = A sin(ωt + φ)`$

</summary>

## :computer::chart_with_upwards_trend: Monitoring
TimeScaleDB integrates seamlessly with Grafana, allowing real-time querying and visualization of sensor data. This enables quick insights into sensor performance, trends, and anomalies. By the same token, it also allows the monitoring of key stats from the database cluster powering up the query of information from the pg_stat_statements and pg_stat_kcache extensions from postgres.

## :cherries: Examples 
(...)

