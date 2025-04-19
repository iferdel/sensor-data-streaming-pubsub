## ROADMAP
### iot-sensor-simluation
- [] protobuf payload instead of json since real scenario
### iot-measurement-ingester
- [] superstreams
### iot-api
- [] tls
- [] sqlc
- [] goose (migrations just for tables. Keep DBA for roles, extensionsn and the alike). Maybe think thourough how would it be in a timescaledb cloud scenario (one database instead of multiple)
- [] sanitization of json
- [] environment variables and production ready
- [] way to communicate between db and sensors regarding sensor state such as sleep/awake, waiting for target,
- [] assign sensor to target endpoint
### iot-cli
- [] api key for auth
- [] set viper for environment variables definition
- [] set bubbletea for beautify the cli tool
- [] TUI in iotctl that enables these kind of changes
- [] one can 'get' the status of all sensors registered, thus being able to see which ones are 'waiting for target association'
### dependencies-db
- [] review of db schema
- [] Add geolocalization
- [] Add target association
- [] add stream graph into iot dashboard
- [] read replica of database so we can separate concerns of database for writing (this service) and reading (any other service)
- [] select * from hypertable_compression_stats('sensor_measurement');
- [] filter writes to disk and buffer flushes to only the measurement insert query + sensorid
- [] track io_timing in on
- [] pg_stat_kcache track cpu usage
- [] track shared blocks for dirtied
### general
- [] add logging for each systems (sent through rabbitmq). Nowadays is just sending sensor simulation logs, this can be improved a lot.
- [] When a new sensor is turned on and no target is assigned (default), no measurement is made (it is registered though), until in iotctl TUI someone assign it to a target with a form. 
- [] with the target assignation step we may need a functional testing of the whole thing.
- [] deadletter exchange and queue for debugging purposes
- [] rabbitmq docs + plugins docs + docs
- [] improvement over nack and ack
### documentation
- [] For architecture diagram: socket svg is not as api as othe symbol could be
- [] Update database schema
---
