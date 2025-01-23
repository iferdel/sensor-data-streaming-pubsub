* add logging 

- For architecture diagram: socket svg is not as api as othe symbol could be
- Update over database schema

* iot-api
* * api key
* * tls
* * sqlc
* * goose (migrations just for tables. Keep DBA for extensionsn and the alike)
* * sanitization of json
* * environment variables and production ready
* * way to communicate between db and sensors regarding sensor state such as sleep/awake, waiting for target,
* * assign sensor to target endpoint

* iot-cli
* * api key with iot-api



	// create api key (further store in database)
	// cli would apply for registration
	// depending on username (for this case) the api key would authorize read-only or all
	// so user registers -> server creates api key and save it in another column from the user table in that specific user's row/record
	// api responds with api key (with https)
	// user store its key locally (it could be done through the cli tool, which could save the api key automatically and refer later on into a dotfile) -- it may also be saved within the cli tool??
	// the user should not share this key
	// anytime the cli tool makes a requests the tool requests includes the api key in the http headers, particularly in "Authorization: Bearer <API_KEY>" header
	// the api key is encrypted (avoid sha-256 bc is fast)
	// the api verifies the api key in every request
	// it extracts the key from the header (or anywhere it is) -> it validates the key with the database -> authorize it or unauthorized (401)
	// role based access control (RBAC)
	// set expiration key (and way to renew the key)
	// database should

-- general
- When a new sensor is turned on and no target is assigned, no measurement is made (it is registered though), until in iotctl TUI someone assign it to a target with a form. 
- From last point, in iotctl one can 'get' the status of all sensors registered, thus being able to see which ones are 'waiting for target association'.

-- iotctl
* way to map to the webserver and all its replicas if any, so using a reverse proxy would be of help. 
* set viper for environment variables definition
* set bubbletea for beautify the cli tool
- TUI in iotctl that enables these kind of changes
