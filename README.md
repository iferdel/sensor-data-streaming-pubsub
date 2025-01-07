# Sensor Data Streaming PubSub

## General Description :telescope:

* (GIF with realtime sensor data being shown up in grafana)*
* (GIF of grafana dashboard showing up scaling up the system with more or less sensors)*
* (GIF showing iotctl behaviour -- maybe with bubbletea implemented already which would beautify the status of running sensors and not running sensors)*
* (GIF showing pods on k8)
* (GIF showing database stats by means of CTE's and pg_stat_statements + pg_stat_kcache)
* General diagram
* -- maybe(GIF showing GPS data from sensors)*


Being ambitious that a design like this is a good starting point to a larger project that would involve real and broader sensor monitoring spectrum with GPS (either with wifi or GSM), in either mobile vehicles or static machinery and why not, maybe in civil infraestructure. 

## Directory Tree :deciduous_tree:
```
.
├── Dockerfile
├── README.md
├── cmd
│   ├── iotctl
│   │   ├── cmd
│   │   │   ├── awake.go
│   │   │   ├── changesamplefrequency.go
│   │   │   ├── root.go
│   │   │   └── sleep.go
│   │   └── main.go
│   ├── sensor
│   │   ├── handlers.go
│   │   └── main.go
│   ├── sensor-logs-ingester
│   │   ├── handlers.go
│   │   └── main.go
│   ├── sensor-measurements-ingester
│   │   ├── handlers.go
│   │   └── main.go
│   └── sensor-registry
│       ├── handlers.go
│       └── main.go
├── compose.yaml
├── go.mod
├── go.sum
├── ideas.md
└── internal
    ├── pubsub
    │   ├── consume.go
    │   └── publish.go
    ├── routing
    │   ├── models.go
    │   └── routing.go
    ├── sensorlogic
    │   ├── awake.go
    │   ├── changesamplefrequency.go
    │   ├── sensordata.go
    │   ├── sensorstate.go
    │   └── sleep.go
    └── storage
        ├── logs.go
        ├── measurements.go
        └── sensors.go

```
## Reason :seedling:

Back in 2020, I worked on **vibration analysis**. My main background back then was in **Mechanical Engineering**, and I took on a role that involved designing sensor installations, performing measurements in the field, and then analyzing them back at the office. I measured various types of mechanical equipment, such as overhead cranes in a mining plant and climate control systems (pumps, cooling towers, air handling units) in Chile’s main airport, as well as civil structures (protected, commercial, and private buildings) subjected to nearby construction or certain physical phenomena.
All of these tasks were performed **in-situ**, which motivated me to consider a more ambitious approach: **remote, real-time monitoring**. Such a system could open up new opportunities for clients, offering continuous insight without requiring on-site personnel.
With that in mind, my goal for this project is to build a full **end-to-end**, real-time monitoring solution.

## Architecture :rabbit: :elephant: :tiger: :whale: :octopus:

*(high-level system diagram to visualize the architecture)*
This solution is based on an **event-driven** architecture using a **pub/sub** pattern at its core, powered by a **distributed system**. The project is intended to be hosted on a **Kubernetes cluster** to ensure high availability and horizontal scaling as needed—for example, if sensor sampling rates increase or if we add more sensors.
The services in question are:
* sensor simulation
* sensor registry
* logs ingester
* measurement ingester
* control iotctl 

Besides this services, to a processing service that stores the data in a database, and finally to a dashboard for visualization.

### Key architectural points :seven::seven::seven::
*Disclaimer: Here we are considering an entire PubSub architecture, but one could conclude that a hybrid architecture for critical low-latency control would also be quite handy. In that case, one would expect using gRPC as the way to communicate between a service that would send direct commands to change behaviour (in a reactive way) not the sensor but to the machine or whatever is behind.*
- **Data Transfer**: The solution is intended to use Protobuf as a data serialization format to match real scenarios with embedded C or C++. However, for the initial setup (POC), the Go encoding/gob serializer will be used.
- **Infrastructure**: This project integrates with my [homelab](https://github.com/iferdel/homelab), which simulates a cloud-like environment on bare metal using TalosOS and GitOps with FluxCD.
- **CI/CD**: For CI/CD, I’m using a private Jenkins server and Docker Hub for image storage, while the GitHub repository hosts the source code. The whole CD would be handled with FluxCD.
- **Secrets**: I’m using Azure Key Vault for secrets in the homelab. 
- **Database**: The solution uses PostgreSQL with [TimeScaleDB](https://www.timescale.com/), an extension optimized for time-series data. In a real scenario, the paid cloud tier would be in use, but for this project I’m storaging everything on bare metal, integrated with CloudNativePG and OpenEBS + Mayastor for storage. 
storage may be ephemeral and with a data retention policy to avoid using too much space on the homelab.
- **Data Management**: TimeScaleDB’s policies handle data expiration and compression, preventing storage overflow and improving performance.
- **Visualization**: Grafana is used for near real-time dashboards, leveraging its querying capabilities to visualize time-series data stored as well as stats from the database itself by means of wrapping the stats from pg_stat_statements and pg_stat_kcache with postgres CTEs and procedures.
- **Alarms**: *...*  
- **Communication Protocols**:
    - *Sensor communication uses MQTT with streaming queues.*
    - *Inter-service communication uses AMQP with RabbitMQ, employing quorum queues.*
    - *Alarm service communication uses gRPC for low-latency communication with the machine where the sensor to affect behaviour*

## Design :art:
'paint on a canvas' - kawhi leonard

Sensors will send:
    - id (serial number)
    - timestamp (with nanosecond precision)
    - measurement (at a variable sample frequency)
    - GPS location (latitude and longitude)
This approach allows flexible scaling of the number of sensors and their data rates.

Sensor will receive (mapped through its id):
    - commands that would affect the sensor behaviour such as sleep, awake and change sample frequency.

## Database Schemas :floppy_disk:

(image of ERD)

## Monitoring :computer:

TimeScaleDB integrates seamlessly with Grafana, allowing real-time querying and visualization of sensor data. This enables quick insights into sensor performance, trends, and anomalies.

## Examples :cherries:

(tmux showing up logs and cmd applying behavioural changes over the sensors.)
(...)

## Thoughts on the Process...

### Simulation

I'm thinking of simulating tens, hundreds and why not, thousands of sensors sending data.
So not only the pubsub is being tested with this high throughput of data, but the database.

### Microservices Considered

In addition to the sensor layer (client), three key microservices will form the core of this architecture:
    Sensor Control Panel: A service (or CLI tool) that interacts with each sensor to adjust operational parameters, such as putting a sensor to sleep or changing its sampling frequency.
    Sensor Data Capture: A horizontally scalable service (utilizing Kubernetes Horizontal Pod Autoscaling) that consumes sensor data and writes it to the database.
    Sensor Data Processing & Alarms: A service that analyzes incoming data to detect patterns or threshold violations and then triggers commands or alarms as needed.

### Pub Sub Compontents to be Used

Exchange, Queues, and Routing Keys:

    Exchange: sensor_streaming
    Queues: sensor_{i} for each sensor i
    Routing Keys:
        sensor_{i}.commands.sleep
        sensor_{i}.commands.change_sample_frequency
        sensor_{i}.data
        sensor_{i}.alarms.trigger

### Pub/Sub Pattern for Each Microservice:

1) Sensor:
    Publisher: Sensor data and acknowledgments after commands or triggers are processed
    Subscriber: Commands and alarms sent from other services
2) Sensor Control Panel:
    Publisher: Commands (with necessary arguments) sent to sensors
    Subscriber: Sensor status updates and command acknowledgments
3) Sensor Data Capture:
    Subscriber: Consumes continuous sensor data streams for persistence in the database
    Sensor Data Processing & Alarms:
        Publisher: Triggers or alerts based on analyzed sensor data
        Subscriber: Receives raw sensor data for processing

