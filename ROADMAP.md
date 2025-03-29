## ROADMAP
### iot-sensor-simluation
- [] protobuf payload instead of json since real scenario
### iot-measurement-ingester
- [x] use [stream plugin](https://www.rabbitmq.com/docs/stream-core-plugin-comparison)
- [x] [stream plugin client in go](https://github.com/rabbitmq/rabbitmq-stream-go-client)
- [] one stream queue per sensor for measurements consumption
- [] single active consumer feature for streams (con instancias de backup esperando por si ese consumer falla)
- [] stream x-max-age parameter
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
- [] Add CTEs for monitoring with timescaledb and pg_stat_statements
### general
- [] add logging for each systems (sent through rabbitmq). Nowadays is just sending sensor simulation logs, this can be improved a lot.
- [] When a new sensor is turned on and no target is assigned (default), no measurement is made (it is registered though), until in iotctl TUI someone assign it to a target with a form. 
- [] with the target assignation step we may need a functional testing of the whole thing.
- [] deadletter exchange and queue for debugging purposes
- [] rabbitmq docs + plugins docs + docs
- [] improvement over nack and ack
- [] measure I/O in stream queue somehow. Nowadays is a bit uncertain since stream queues are stored on disk as a append only logs
### documentation
- [] For architecture diagram: socket svg is not as api as othe symbol could be
- [] Update database schema
---
