# General Description

(GIF with realtime sensor data being shown up in grafana)
(GIF of grafana dashboard showing up scaling up the system with more or less sensors)
(GIF showing GPS data from sensors)

# Reason
In 2019 I worked with vibrational analysis. I studied Mechanical Engineering and got into this place where they needed someone that would perform the measurements and then analyse them. I did measure mechanical equipment such as overhead cranes, climate equipment for the main airport in Chile (pumps, cooling towers, air handling units), aswell as civil structures. Everyting was thought on being measured in-situ. Which is one of the motivations of this project, since I think a more ambitious and profitable solution would involve remote monitoring. This would open up doors for new requests for clients. 

# Architecture

# Design

# Database Schemas

# Monitoring

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


