# General Description

(GIF with realtime sensor data being shown up in grafana)
(GIF of grafana dashboard showing up scaling up the system with more or less sensors)
(GIF showing GPS data from sensors)

# Reason
Back in 2020 I worked with vibrational analysis. 
I studied Mechanical Engineering and got into this place where they needed someone that would design for the installation of the sensors and perform the measurements, and then analyse them back at office. I did measure mechanical equipment such as overhead cranes in a mining plant, climate equipment for the main airport in Chile (pumps, cooling towers, air handling units), aswell as civil structures (protected buildings, commercial buildings, private buildings) subjected to nearby constructions or certain physical phenomenas. Everyting was thought on being measured in-situ. Which is one of the motivations for this project, since I think a more ambitious and profitable solution would involve remote monitoring. This would have open up interesting doors for new requests for clients.

# Architecture

sensor layer
The whole solution is based on an event driven architecture, using PubSub pattern being powered up by a distributed system. This project is intended to be hosted in a kubernetes cluster which would allow high availability and horizontal pod scaling whenever needed (for example, with the increasing of sample rate of the sensors or a scale up process of plugging more sensors into the system)
The data is transfered using protobuff, but the demo is being constructed over Gob serializer provided by the standard library encoding/gob
I have been working on a (homelab)[https://github.com/iferdel/homelab]. This homelab is based on different environments and bare metal using talosOS to work on immutable cluster, with similitudes to what one may expect in a cloud based solution. This whole homelab kubernetes configurations is powered up by GitOps using FluxCD.
For this repo my CI is being linked between this github repo (I ommited any intention to have a private git server for seamless demostration of this project), jenkins server running inside the homelab, and dockerhub for image storage.
I am using Azure Secrets within the homelab for handling the database instance. In this case, I am using TimeScaleDB as the database to store sensor data. In a real scenario I would go all into the cloud based solution, but for this project I am using a storage from within the bare metal cluster using CloudNativePG and OpenEBS + Mayastor
To visualize the information being retrieved from the sensors I am using Grafana, taking advantage of the querying options for near real-time monitoring.
The data being saved in the database is thought to be expiring with expiring policies from timescaleDB, so I wont cope the entire disk with data. On the other hand I am using compression directly over the database.

# Design

I want sensors that sends information such as 
timestamp, measurement (based on a variable sample frequency)
gps location latitude and longitude

# Database Schemas

# Monitoring

TimeScale allows me to query the database from within Grafana

# Examples

# Thoughts on the Process...
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


