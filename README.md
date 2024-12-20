# Sensor Data Streaming PubSub

## General Description

*(GIF with realtime sensor data being shown up in grafana)*
*(GIF of grafana dashboard showing up scaling up the system with more or less sensors)*
*(GIF showing GPS data from sensors)*

## Directory Tree
```
.
├── Dockerfile
├── README.md
├── cmd
│   ├── client
│   │   ├── handlers.go
│   │   └── main.go
│   └── server
│       ├── cmd
│       │   ├── awake.go
│       │   ├── changesamplefrequency.go
│       │   ├── root.go
│       │   └── sleep.go
│       └── main.go
├── compose.yaml
├── go.mod
├── go.sum
└── internal
    ├── pubsub
    │   ├── consume.go
    │   └── publish.go
    ├── routing
    │   ├── models.go
    │   └── routing.go
    └── sensorlogic
        ├── awake.go
        ├── changesamplefrequency.go
        ├── sensordata.go
        ├── sensorstate.go
        └── sleep.go
```
## Reason

Back in 2020, I worked on **vibration analysis**. My main background back then was in **Mechanical Engineering**, and I took on a role that involved designing sensor installations, performing measurements in the field, and then analyzing them back at the office. I measured various types of mechanical equipment, such as overhead cranes in a mining plant and climate control systems (pumps, cooling towers, air handling units) in Chile’s main airport, as well as civil structures (protected, commercial, and private buildings) subjected to nearby construction or certain physical phenomena.
All of these tasks were performed **in-situ**, which motivated me to consider a more ambitious approach: **remote, real-time monitoring**. Such a system could open up new opportunities for clients, offering continuous insight without requiring on-site personnel.
With that in mind, my goal for this project is to build a full **end-to-end**, real-time monitoring solution: from sensors streaming data (*simulated*) to a server that can adjust their behavior, to a processing service that stores the data in a database, and finally to a dashboard for visualization.

## Architecture

*(high-level system diagram to visualize the architecture)*
This solution is based on an **event-driven** architecture using a **pub/sub** pattern, powered by a **distributed system**. The project is intended to be hosted on a **Kubernetes cluster** to ensure high availability and horizontal scaling as needed—for example, if sensor sampling rates increase or if we add more sensors.

### Key architectural points:
- **Data Transfer**: The solution is intended to use Protobuf as a data serialization format to match real scenarios with embedded C or C++. However, for the initial setup (POC), the Go encoding/gob serializer will be used.
- **Infrastructure**: This project integrates with my [homelab](https://github.com/iferdel/homelab), which simulates a cloud-like environment on bare metal using TalosOS and GitOps with FluxCD.
- **CI/CD and Images**: For CI/CD, I’m using Jenkins inside the homelab and Docker Hub for image storage, while the GitHub repository hosts the source code (pulling the latest image by means of FluxCD source controllers).
- **Secrets**: I’m using Azure Secrets in the homelab for secure database credentials. 
- **Database**: PostgreSQL is used with TimeScaleDB, an extension optimized for time-series data. In a real scenario, this would be fully cloud-based, but for this project, I’m running it on bare metal, integrated with CloudNativePG and OpenEBS + Mayastor for storage.
- **Data Management**: TimeScaleDB’s policies handle data expiration and compression, preventing storage overflow and improving performance.
- **Visualization**: Grafana is used for near real-time dashboards, leveraging its querying capabilities to visualize time-series data stored in TimeScaleDB as well as stats from the database itself by means of wrapping the stats from pg_stat_statements and pg_stat_kcache with postgres CTEs and procedures.
- **Communication Protocols**:
    - *Sensor communication uses MQTT with streaming queues.*
    - *Inter-service communication uses AMQP with RabbitMQ, employing both durable and transient queues.*

## Design

Sensors will send:
    - timestamp (with nanosecond precision)
    - measurement (at a variable sample frequency)
    - GPS location (latitude and longitude)
This approach allows flexible scaling of the number of sensors and their data rates.

## Database Schemas

(image of ERD)

## Monitoring

TimeScaleDB integrates seamlessly with Grafana, allowing real-time querying and visualization of sensor data. This enables quick insights into sensor performance, trends, and anomalies.

## Examples

(tmux showing up logs and cmd applying behavioural changes over the sensors.)
(...)

## Thoughts on the Process...

### Simulation

I'm thinking of simulating tens, hundreds and why not, thousands of sensors sending data.

### Microservices Considered

In addition to the sensor layer, three key microservices will form the core of this architecture:
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

