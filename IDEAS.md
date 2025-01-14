-- general
- TUI in iotctl that enables these kind of changes
- When a new sensor is turned on and no target is assigned, no measurement is made (it is registered though), until in iotctl TUI someone assign it to a target with a form. 
- From last point, in iotctl one can 'get' the status of all sensors registered, thus being able to see which ones are 'waiting for target association'.

-- iotctl
* this cli tool is intended to be run remotely in any local machine with access to the cluter that contains the project.
* in that regard, having remote and direct access to the broker and the database is nonsense.
* that's why this idea of interacting from the cli to sensors should be split in two, 
* one dedicated to send instructions with an auth feature and another that receives these requests through https,
* process, validate it and send a response to the cli remote user aswell as a message(s) to the msg broker in order to fulfill the needs.
* create a dedicated webserver service to get an endpoint to do post and get requests (commands) instead of sending messages directly to the broker.
* this way is easier to login with auth tokens
* way to map to the webserver and all its replicas if any, so using a reverse proxy would be of help. 
* This would handle the command requests based on login needs and then sends messages to the broker.
* set viper for environment variables definition
* set bubbletea for beautify the cli tool
