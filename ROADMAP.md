## ROADMAP

### iot-sensor-simluation
- [ ] simulación empieza randomly primero pickea un tipo de sensor (T°, humedad, vibraciones, etc. Todos con GPS) y luego según eso selecciona con qué sample frequency comienza la medición, ahí nos ahorramos esa environment variable del compose. Ej, vibraciones tiene Fs desde 3.000 a 25.000, T° tiene Fs de 60, y así. Mismo tema con serial number, tiene que generarse ese serial number de forma aleatoria en el init y el sistema lo reconoce con el header de qué tipo de sensor es... cosa de que se guarde con ese 'id, tipo'.
- [ ] protobuf payload instead of json since real scenario
- [ ] gps simulation
- [ ] sine waves for vibrational measurements
- [ ] randomly create simulation of either, T°, humidity, vibration, odometer
### iot-measurement-ingester
- [ ] superstreams
- [ ] este servicio maneja muchos requests a la base de datos (real time monitoring) así que caching es clave y debería estar pensado desde el momento inicial.
- [ ] concurrency (goroutines)
### iot-log-ingester
- [ ] File I/O per log message - revisar si el log file se está abriendo y cerrando por cada log message. Si ese es el caso debería haber una manera más 'elegante' de hacerlo.
### iot-api
- [ ] tls
- [ ] sqlc
- [ ] goose (migrations just for tables. Keep DBA for roles, extensionsn and the alike). Maybe think thourough how would it be in a timescaledb cloud scenario (one database instead of multiple)
- [ ] sanitization of json
- [ ] environment variables and production ready
- [ ] way to communicate between db and sensors regarding sensor state such as sleep/awake, waiting for target,
- [ ] assign sensor to target endpoint
- [ ] este servicio maneja muchos requests a la base de datos (real time monitoring) así que caching es clave y debería estar pensado desde el momento inicial.
- [ ] está interesante este patrón ... creo que es del texto de design patterns de refactoring guru
```
  // Common in HTTP clients
  type Client[T any] interface {
      Get(url string) (*T, error)
      Post(url string, body T) error
  }

// Database repositories
type Repository[T Entity] interface {
    Save(T) error
    FindByID(id string) (T, error)
}

```
### iot-cli
- [ ] api key for auth
- [ ] set viper for environment variables definition
- [ ] set bubbletea for beautify the cli tool
- [ ] TUI in iotctl that enables these kind of changes
- [ ] one can 'get' the status of all sensors registered, thus being able to see which ones are 'waiting for target association'
- [ ] está interesante este patrón ... creo que es del texto de design patterns de refactoring guru
```
  // Common in HTTP clients
  type Client[T any] interface {
      Get(url string) (*T, error)
      Post(url string, body T) error
  }

// Database repositories
type Repository[T Entity] interface {
    Save(T) error
    FindByID(id string) (T, error)
}

```
### dependencies-db
- [ ] review of db schema
- [ ] Add geolocalization
- [ ] Add target association
- [ ] add stream graph into iot dashboard
- [ ] read replica of database so we can separate concerns of database for writing (this service) and reading (any other service)
- [ ] select * from hypertable_compression_stats('sensor_measurement');
- [ ] filter writes to disk and buffer flushes to only the measurement insert query + sensorid
- [ ] track io_timing in on
- [ ] pg_stat_kcache track cpu usage
- [ ] track shared blocks for dirtied
- [ ] pg autotune AI agent para monitorear pg cluster
- [ ] couple of 'hacking' series of vids from [these guys](https://www.youtube.com/@PostgresTV)
- [ ] play aroudnb woprk_maintenance_mem https://postgresqlco.nf/doc/en/param/maintenance_work_mem/
### dependencies-grafana / prometheus
- [ ] monitoreio para exhausting vertical scaling antes de complejizar con scaling horizontal. Ej. I/O is quite high, write operations / read operations scale.
- [ ] graph in grafana the dead tuples to monitor autovaccum behaviour
```
  SELECT 
    NOW() as time,
    relname,
    n_dead_tup as dead_tuples,
    n_live_tup as live_tuples,
    n_dead_tup::float / NULLIF(n_live_tup, 0) * 100 AS dead_percentage
FROM pg_stat_user_tables
WHERE schemaname = 'public' 
  AND relname LIKE '%sensor_measurement%'
ORDER BY n_dead_tup DESC;

SELECT 
    NOW() as time,
    schemaname,
    relname,
    last_autovacuum,
    last_autoanalyze,
    autovacuum_count,
    autoanalyze_count
FROM pg_stat_user_tables
WHERE relname LIKE '%sensor_measurement%'
ORDER BY last_autovacuum DESC;
```
- [ ] in grafana agregar versioning de dependencias actuales. Ej. rabbitmq v.xxx golang v1.25.xxx timescaledb v1.1.1
- [ ] usar grafana scenes para crear custom dashboard. Creo que usa embedded js... ver docs
- [ ] drilldowns (averiguar si se peude hacer con el open source o es sólo para grafana cloud...) en grafana dashboard (para navegar de una vista general a una personalizada de un sensor en particular, por ejemplo) -> es al final un link a otro dashboard
- [ ] buscar en data sources-.. insights. Creo que esto es sólo para grafana cloud. Probar acá. Al final menciona info como 'queries' stats generales.
- [ ] use grafana.org/dashboard
- [ ] play.grafana.org -> many examples...
- [ ] usar en homelab el grafana dashboard público para que puedan ingresar como una especie de 'portafolio' de ese proyecto en particular... concienzudamente pensar sobre authorization y esas cosas...
- [ ] usar grafana play...
- [ ] usar transofrmaciones en grafana para, por ejemplo filtrar los últimos logs... aunque esto en vdd puede ser desde el query en sí... hay otras opciones para implementar
### general
- [ ] [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)
- [ ] [semver](https://semver.org/lang/es/)
- [ ] [keep a changelog](https://keepachangelog.com/en/1.1.0/)
- [ ] [devops 101](https://www.youtube.com/watch?v=QDpVgt1zn2M&t=857s)
- [ ] ELT (not ETL) con dbt podría ser útil para la experiencia del datawarehouse, necesitaría la tech de datawarehouse.
- [ ] usar lo aprendido (y librerías) utilizadas en tcp to http, httpclient, http server, web crawler, go course de bootdev.
- [ ] add logging for each systems (sent through rabbitmq). Nowadays is just sending sensor simulation logs, this can be improved a lot.
- [ ] When a new sensor is turned on and no target is assigned (default), no measurement is made (it is registered though), until in iotctl TUI someone assign it to a target with a form. 
- [ ] with the target assignation step we may need a functional testing of the whole thing.
- [ ] deadletter exchange and queue for debugging purposes
- [ ] rabbitmq docs + plugins docs + docs
- [ ] improvement over nack and ack
- [ ] pgroute for analysis of geolocation with [graph capabilities](https://www.timescale.com/learn/postgresql-extensions-pgrouting)
- [ ] AMQP channel creation per HTTP request - esta operación debe ser pooleada en vez de abrir nuevos channels por cada http request. Revisar este punto en todos los servicios que hacen el publish o subscribe.
- [ ] user %w en Errorf formatting para mejor output del error message
- [ ] [Architect LARGE software projects: minuto 43.00, 45.00, 52.00, 1.14.00](https://www.youtube.com/watch?v=sSpULGNHyoI&t=73s)
- [ ] corroborar uso de estos statements:
```
      - [ ] Don't communicate by sharing memory, share memory by communicating.
      - [ ] Concurrency is not parallelism.
      - [ ] Channels orchestrate; mutexes serialize.
      - [ ] The bigger the interface, the weaker the abstraction.
      - [ ] Make the zero value useful.
      - [ ] interface{} says nothing.
      - [ ] Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.
      - [ ] A little copying is better than a little dependency.
      - [ ] Syscall must always be guarded with build tags.
      - [ ] Cgo must always be guarded with build tags.
      - [ ] Cgo is not Go.
      - [ ] With the unsafe package there are no guarantees.
      - [ ] Clear is better than clever.
      - [ ] Reflection is never clear.
      - [ ] Errors are values.
      - [ ] Don't just check errors, handle them gracefully.
      - [ ] Design the architecture, name the components, document the details.
      - [ ] Documentation is for users.
      - [ ] Don't panic.
```
- [ ] ver uso de iotas, iota no es un enum... a pesar de que lo parezca. Una iota es sólo una secuencia de números.
- [ ] ver uso de type definitions. Ej.
```
                          type sendingChannel string

                          const (
                              Email sendingChannel = "email"
                              SMS   sendingChannel = "sms"
                              Phone sendingChannel = "phone"
                          )

                          func sendNotification(ch sendingChannel, message string) {
                              // send the message
                          }

                          // The following checkPermission(Admin) will throw an error unless it is called as
                            // checkPermission(perm(Admin))
                          type perm string

                          const (
                              Read  perm = "read"
                              Write perm = "write"
                              Exec  perm = "execute"
                          )

                          var Admin = "admin"
                          var User = perm("user")

                          func checkPermission(p perm) {
                              // check the permission
                          }
```
- [ ] revisar si uso un empty struct como placeholder
- [ ] revisar uso de variables variádicas
- [ ] revisar uso de maps
- [ ] identificar thread safety en servicios
- [ ] revisar cuidadosamente el uso de genéricos. Se pueden usar custom constraints para restringir el [T any] en algo así como [T customInterface(as type constraint)]
- [ ] ver uso de interfaces. Apoyarme de texto o ejemplos.
- [ ] ver uso de interfaces con genéricos, en particualr [parametric constraints](https://www.boot.dev/lessons/4a9635d1-9bd9-40b4-81b7-d3662aa3889c), puede servir para handling different response types en APIs, middleware, o bien diferentes entity types mientras se trabaja con una DB. ejemplo, se usa mucho en java, por ejemplo, entonces es útil entenderlos porque ayuda al enterprise codebases en go. Tambien se usa en rust con el nombre de 'traits with associated types'. Sirve hacer distinción entre parameter types y s. Sirve hacer distinción entre parameter types y specific typespecific types:
- [ ] ver donde aplicar mutexes
- [ ] ver si puedo mejorar tratamiento de channels con select
- [ ] usar nuevos conocimientos de pointers para configuraciones y demáses
- [ ] usar nuevos conocimientos de currying (pseudo closure) para abstraer funciones. No así métodos
- [ ] cuando se tenga que retornar en base a un booleano se puede retornar la condición en vez de todo el if statement y luego un return true o false.
- [ ] revisión del uso de make statements (y buffered channels)
- [ ] revisión uso de anonymous functions
- [ ] revisión de uso de anonymous structs
- [ ] revisión de cuando se usa nested o embedded structsits is always prefered embedded structs rather than
- [ ] accept interfaces return struct
### documentation
- [ ] For architecture diagram: socket svg is not as api as othe symbol could be
- [ ] Update database schema
- [ ] documentar plan protocolo de monitoreio para exhausting vertical scaling antes de complejizar con scaling horizontal. Ej. I/O is quite high, write operations / read operations scale.
- [ ] documentar Cloud native solutions are awesome and one can get terabytes of memory and hundreds of cpus
