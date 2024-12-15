# Sensor Data Streaming PubSub

## General Description

*(GIF with realtime sensor data being shown up in grafana)*
*(GIF of grafana dashboard showing up scaling up the system with more or less sensors)*
*(GIF showing GPS data from sensors)*

## Reason

Back in 2020, I worked on **vibration analysis**. My main background back then was in **Mechanical Engineering**, and I took on a role that involved designing sensor installations, performing measurements in the field, and then analyzing them back at the office. I measured various types of mechanical equipment, such as overhead cranes in a mining plant and climate control systems (pumps, cooling towers, air handling units) in Chile’s main airport, as well as civil structures (protected, commercial, and private buildings) subjected to nearby construction or certain physical phenomena.
All of these tasks were performed **in-situ**, which motivated me to consider a more ambitious approach: **remote, real-time monitoring**. Such a system could open up new opportunities for clients, offering continuous insight without requiring on-site personnel.
With that in mind, my goal for this project is to build a full **end-to-end**, real-time monitoring solution: from sensors streaming data (*simulated*) to a server that can adjust their behavior, to a processing service that stores the data in a database, and finally to a dashboard for visualization.

## Architecture

*(high-level system diagram to visualize the architecture)*
This solution is based on an **event-driven** architecture using a **pub/sub** pattern, powered by a **distributed system**. The project is intended to be hosted on a **Kubernetes cluster** to ensure high availability and horizontal scaling as needed—for example, if sensor sampling rates increase or if we add more sensors.

### Key architectural points:
- **Data Transfer**: The solution is intended to use Protobuf as a data serialization format. However, for the initial setup (POC), the Go encoding/gob serializer will be used.
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

I want sensors that sends information such as 
timestamp, measurement (based on a variable sample frequency)
gps location latitude and longitude

## Database Schemas

## Monitoring

TimeScale allows me to query the database from within Grafana

## Examples

## Thoughts on the Process...
* using gob as serializer for enconding and decoding messages fits only for the demo, but lather on I'd expect to have protobuff to match a real scenario with sensors working along with embedded C or C++.
* as a matter of choice, the database that will handle the sensor data is going to be postgreSQL with the time-series extension TimeScaleDB. This facilitates a lot along with other postgres extensions for monitoring real time data and statistics such as from pg_stat_statements and pg_stats_kcache. Also, I see a motivation on using a relational database since each sensor would have certain pattern for entities that would match pretty well in this scenario.
* I'm using rabbitMQ as the message broker since its flexibility fits well enough to treat the streaming of data from the sensors using streaming queues along with MQTT, and by the same token, allowing other microservices to work with the same instance of rabbitMQ but with AMQP and standard queues.
* One exchange is in use, one durable queue per sensor since we are expecting data rates on between 1.000 to 25.000 Hz.
* I'm thinking of simulating tens, hundreds and why not, thousands of sensors sending data.
* Each sensor will publish their location and measurements with a timestamp sensible to nanoseconds as our maximum sample frequency is aroung 25.000 Hz (1/25.000 = 4 "mu"s)
* Besides the sensor layer, there are three microservices in mind.
1) sensor control panel: CMD that interacts with each sensor on their operational behaviour. Like sleep, changes on their sample frequency, etc.
2) sensor data capture: Horizontal Pod Autoscaled service which will consume the sensor data and insert it into the database.
3) sensor data processing alarms: Service that triggers behaviour of certain sensors based on patterns such as getting into a threshold (saturation) multiple times over a controlled period of time.

The pattern for exchanges, queues and routing keys are as follow:
exchange: sensor_streaming
queues: sensor_{i} where i is the serial number of the sensor
routing keys: 
    sensor_{i}.commands.sleep
    sensor_{i}.commands.change_sample_frequency
    sensor_{i}.data
    sensor_{i}.alarms.trigger

pubsub pattern per element (microservice):
    sensor: 
        - publisher: 
            - of their captured data
            - of return behaviour after commands/triggers were received
        - subscriber: 
            - to commands and triggers from alarms
    sensor control plane:
        - publisher: 
            - of their commands with arguments
        - subscriber: 
            - to sensor state based on return values from commands
    sensor data capture:
        - subscriber:
            - to data streaming from sensors
    sensor data processing alarms:
        - publisher:
            - of their triggers based on processed data


