# Design Patterns Reference - Sensor Data Streaming System

**Repository**: sensor-data-streaming-pubsub
**Purpose**: Comprehensive reference document mapping all design patterns in the codebase
**Last Updated**: 2025-11-17

> This document serves as a "wisdom reference" for understanding the architectural decisions and design patterns already implemented in this repository. Use this when onboarding, refactoring, or adding new features to maintain consistency.

---

## Table of Contents

1. [Architectural Patterns](#architectural-patterns)
2. [Creational Patterns](#creational-patterns)
3. [Structural Patterns](#structural-patterns)
4. [Behavioral Patterns](#behavioral-patterns)
5. [Concurrency Patterns](#concurrency-patterns)
6. [Data Patterns](#data-patterns)
7. [HTTP/Response Patterns](#httpresponse-patterns)
8. [Configuration Patterns](#configuration-patterns)

---

## Architectural Patterns

### 1. Hexagonal Architecture (Ports & Adapters)

**What it is**: The application is organized with a clear separation between business logic (core) and external concerns (adapters). The core doesn't know about databases, message brokers, or protocols.

**Directory Structure**:

```
internal/
├── sensorlogic/        # CORE (Business Logic - Domain Layer)
│   ├── sensor.go       # Domain entities and state management
│   ├── sensormeasurements.go  # Measurement business logic
│   ├── commands.go     # Command definitions
│   └── ...
├── pubsub/             # PORT (Message Broker Interface)
│   ├── consume.go      # Abstract consumption interface
│   └── publish.go      # Abstract publishing interface
├── storage/            # PORT (Database Interface)
│   ├── db.go          # Database connection abstraction
│   ├── sensors.go     # Sensor repository
│   └── measurements.go # Measurement repository
└── routing/            # PORT (Message Routing Configuration)
    ├── models.go       # DTOs (Data Transfer Objects)
    └── routing.go      # Routing keys and queue names

cmd/
├── sensor-simulation/  # ADAPTER (MQTT/AMQP Producer)
├── sensor-registry/    # ADAPTER (AMQP Consumer -> PostgreSQL)
├── sensor-measurements-ingester/  # ADAPTER (Stream Consumer -> PostgreSQL)
├── sensor-logs-ingester/  # ADAPTER (AMQP Consumer -> File System)
├── iot-api/           # ADAPTER (HTTP -> AMQP)
└── iotctl/            # ADAPTER (CLI -> HTTP)
```

**Specific Examples**:

**CORE** - Business Logic (`internal/sensorlogic/sensormeasurements.go:111-140`):
```go
// Pure business logic - doesn't know HOW data is fetched or stored
func HandleMeasurements(ctx context.Context, db *storage.DB, dtos []routing.SensorMeasurement) error {
    sensorMap, err := db.GetSensorIDBySerialNumberMap(ctx)  // Uses PORT interface
    if err != nil {
        return fmt.Errorf("failed to fetch sensor IDs: %v", err)
    }

    records := make([]storage.SensorMeasurementRecord, len(dtos))
    for i, dto := range dtos {
        sensorID, exists := sensorMap[dto.SerialNumber]
        if !exists {
            return fmt.Errorf("sensor serial number not found: %s", dto.SerialNumber)
        }
        records[i] = storage.SensorMeasurementRecord{
            Timestamp:   dto.Timestamp,
            SensorID:    sensorID,
            Measurement: dto.Value,
        }
    }

    return db.BatchArrayWriteMeasurement(ctx, records)  // Uses PORT interface
}
```

**PORT** - Database Interface (`internal/storage/db.go:16-40`):
```go
// PORT - Abstracts database operations
type DB struct {
    pool *pgxpool.Pool  // Concrete implementation hidden
}

func NewDBPool(connString string) (*DB, error) {
    ctx := context.Background()
    dbpool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %v", err)
    }
    return &DB{pool: dbpool}, nil
}

func (db *DB) Close() {
    db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
    return db.pool.Ping(ctx)
}
```

**PORT** - Message Broker Interface (`internal/pubsub/consume.go:77-100`):
```go
// PORT - Abstract subscription interface (protocol-agnostic)
func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange, queueName, key string,
    queueDurability QueueDurability,
    queueType QueueType,
    handler func(T) AckType,
) error {
    return subscribe[T](conn, exchange, queueName, key, queueDurability, queueType, handler,
        func(data []byte) (T, error) {
            var target T
            err := json.Unmarshal(data, &target)
            return target, err
        },
    )
}
```

**ADAPTER** - HTTP to AMQP (`cmd/iot-api/main.go:15-36`):
```go
// ADAPTER - Connects HTTP world to messaging world
type apiConfig struct {
    ctx        context.Context
    rabbitConn *amqp.Connection
    db         *storage.DB
}

func NewApiConfig() (*apiConfig, error) {
    ctx := context.Background()
    conn, err := amqp.Dial(routing.RabbitConnString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }
    db, err := storage.NewDBPool(storage.PostgresConnString)
    return &apiConfig{
        ctx:        ctx,
        rabbitConn: conn,
        db:         db,
    }, nil
}
```

**Why Hexagonal?**:
- ✅ Business logic (`internal/sensorlogic/`) has zero dependencies on external frameworks
- ✅ Can swap PostgreSQL for another DB by only changing `internal/storage/` implementation
- ✅ Can swap RabbitMQ for Kafka by only changing `internal/pubsub/` implementation
- ✅ Tests can mock `storage.DB` interface without touching real database

---

### 2. Event-Driven Architecture (Pub/Sub)

**What it is**: Services communicate asynchronously via message broker (RabbitMQ). Publishers don't know about subscribers.

**Message Flow Diagram**:

```
[sensor-simulation]
       ↓ MQTT (publish)
    [RabbitMQ]
       ↓ Stream (consume)
[sensor-measurements-ingester] → [PostgreSQL]

[sensor-simulation]
       ↓ AMQP (publish)
    [RabbitMQ]
       ↓ AMQP (consume)
[sensor-logs-ingester] → [File System]

[sensor-simulation]
       ↓ AMQP (publish)
    [RabbitMQ]
       ↓ AMQP (consume)
[sensor-registry] → [PostgreSQL]

[iot-api]
       ↓ AMQP (publish)
    [RabbitMQ]
       ↓ AMQP (consume)
[sensor-simulation] → [Handle Command]
```

**Publisher Example** (`cmd/sensor-simulation/main.go:145-153`):
```go
// Publishes sensor registration event - doesn't know who consumes it
pubsub.PublishGob(
    publishCh,                // channel
    routing.ExchangeTopicIoT, // exchange
    fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created", // routing key
    routing.Sensor{
        SerialNumber:    serialNumber,
        SampleFrequency: sampleFrequency,
    },
)
```

**Subscriber Example** (`cmd/sensor-registry/main.go` - assuming similar pattern):
```go
// Subscribes to sensor registration events - doesn't know who publishes
err = pubsub.SubscribeGob(
    conn,
    routing.ExchangeTopicIoT,
    routing.QueueSensorRegistry,
    fmt.Sprintf(routing.KeySensorRegistryFormat, "*")+"."+"#",
    pubsub.QueueDurable,
    pubsub.QueueClassic,
    handlerSensorRegistry(ctx, db),
)
```

**Key Files**:
- Exchange config: `internal/routing/routing.go:7-9`
- Queue config: `internal/routing/routing.go:12-20`
- Routing keys: `internal/routing/routing.go:23-31`

**Benefits**:
- ✅ Loose coupling: Services don't need direct references
- ✅ Scalability: Multiple consumers can process messages in parallel
- ✅ Resilience: If one service is down, messages queue up

---

### 3. Microservices Architecture

**What it is**: System decomposed into 6 independent services, each with single responsibility.

**Services**:

| Service | Directory | Responsibility | Port | Protocol |
|---------|-----------|---------------|------|----------|
| `iot-api` | `cmd/iot-api/` | HTTP API gateway for sensor commands | 8080 | HTTP → AMQP |
| `sensor-simulation` | `cmd/sensor-simulation/` | Simulates IoT sensors, publishes data | N/A | MQTT, AMQP |
| `sensor-registry` | `cmd/sensor-registry/` | Handles sensor enrollment to DB | N/A | AMQP → PostgreSQL |
| `sensor-measurements-ingester` | `cmd/sensor-measurements-ingester/` | Persists measurements to TimescaleDB | N/A | Stream → PostgreSQL |
| `sensor-logs-ingester` | `cmd/sensor-logs-ingester/` | Aggregates and writes logs | N/A | AMQP → File System |
| `iotctl` | `cmd/iotctl/` | CLI for managing sensors | N/A | CLI → HTTP |

**Each service has**:
- Own `main.go` entry point
- Independent deployment (can be containerized separately)
- Single responsibility
- Clear communication contracts (DTOs in `internal/routing/models.go`)

---

## Creational Patterns

### 1. Factory Pattern

**What it is**: Factory methods encapsulate complex object construction with dependencies.

#### Factory: Database Connection Pool

**Location**: `internal/storage/db.go:20-31`

```go
// Factory method creates and configures DB with connection pool
func NewDBPool(connString string) (*DB, error) {
    ctx := context.Background()

    dbpool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %v", err)
    }

    return &DB{
        pool: dbpool,
    }, nil
}
```

**Usage** (`cmd/iot-api/main.go:29`):
```go
db, err := storage.NewDBPool(storage.PostgresConnString)
```

**Why Factory?**: Hides connection pool configuration, retry logic, and initialization complexity.

#### Factory: API Configuration

**Location**: `cmd/iot-api/main.go:21-36`

```go
// Factory creates apiConfig with all dependencies wired
func NewApiConfig() (*apiConfig, error) {
    ctx := context.Background()

    conn, err := amqp.Dial(routing.RabbitConnString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }

    db, err := storage.NewDBPool(storage.PostgresConnString)

    return &apiConfig{
        ctx:        ctx,
        rabbitConn: conn,
        db:         db,
    }, nil
}
```

**Usage** (`cmd/iot-api/main.go:49`):
```go
apiCfg, err := NewApiConfig()
if err != nil {
    log.Fatal(err)
}
defer apiCfg.rabbitConn.Close()
defer apiCfg.db.Close()
```

**Why Factory?**: Single function creates complex object with RabbitMQ connection, DB pool, and context.

#### Factory: Sensor State

**Location**: `internal/sensorlogic/sensor.go:23-36`

```go
// Factory creates SensorState with all channels initialized
func NewSensorState(serialNumber string, SampleFrequency float64) *SensorState {
    return &SensorState{
        Sensor: Sensor{
            SerialNumber: serialNumber,
        },
        LogsInfo:                  make(chan string, 1),
        LogsWarning:               make(chan string, 1),
        LogsError:                 make(chan string, 1),
        SampleFrequency:           SampleFrequency,
        SampleFrequencyChangeChan: make(chan float64, 1),
        IsSleep:                   false,
        IsSleepChan:               make(chan bool, 1),
    }
}
```

**Usage** (`cmd/sensor-simulation/main.go:93`):
```go
sensorState := sensorlogic.NewSensorState(serialNumber, sampleFrequency)
```

**Why Factory?**: Ensures all 5 channels are properly initialized (easy to forget one).

#### Factory: Sensor Cache

**Location**: `internal/sensorlogic/sensormeasurements.go:21-34`

```go
// Factory creates cache and performs initial load from DB
func NewSensorCache(ctx context.Context, db *storage.DB) (*SensorCache, error) {
    cache := &SensorCache{
        db:      db,
        mapping: make(map[string]int),
    }

    // initial fetch (refresh is stated for refreshing, but it could be used here to state it on init)
    err := cache.refresh(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed initial sensor cache load: %v", err)
    }

    return cache, nil
}
```

**Why Factory?**: Ensures cache is populated on construction, not lazily (fail-fast).

---

### 2. Singleton Pattern (Connection Pools)

**What it is**: Single instance of expensive resources (DB connections, message broker connections) shared across application.

**Database Pool Singleton** (`internal/storage/db.go:16-18`):
```go
type DB struct {
    pool *pgxpool.Pool  // Single connection pool instance
}
```

**Created once** (`cmd/iot-api/main.go:49-54`):
```go
apiCfg, err := NewApiConfig()  // Creates ONE DB pool
if err != nil {
    log.Fatal(err)
}
defer apiCfg.rabbitConn.Close()
defer apiCfg.db.Close()
```

**Passed to all handlers** (`cmd/iot-api/main.go:60-69`):
```go
// All handlers share the same apiCfg (and thus same DB pool)
router.HandleFunc("GET /api/v1/sensors", apiCfg.handlerSensorsRetrieve)
router.HandleFunc("GET /api/v1/sensors/{sensorSerialNumber}", apiCfg.handlerSensorsGet)
router.HandleFunc("PUT /api/v1/sensors/{sensorSerialNumber}/sleep", apiCfg.handlerSensorsSleep)
// ... all use apiCfg.db (same instance)
```

**Why Singleton?**:
- ✅ Connection pooling: Reuse DB connections instead of creating new ones per request
- ✅ Resource efficiency: One pool manages connections for entire app
- ⚠️ Note: Not a "strict" singleton (no global state), but single instance per service

---

### 3. Builder Pattern

**What it is**: Fluent interface for constructing complex configuration objects.

#### MQTT Client Builder

**Location**: `cmd/sensor-simulation/main.go:32-42`

```go
func MQTTCreateClientOptions(clientId, raw string) *mqtt.ClientOptions {
    uri, _ := url.Parse(raw)
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
    opts.SetUsername(uri.User.Username())
    password, _ := uri.User.Password()
    opts.SetPassword(password)
    opts.SetClientID(clientId)

    return opts
}
```

**Usage** (`cmd/sensor-simulation/main.go:52`):
```go
mqttOpts := MQTTCreateClientOptions("publisher", routing.RabbitMQTTConnString)
mqttClient := mqtt.NewClient(mqttOpts)
```

**Why Builder?**: MQTT options have many optional fields - builder pattern makes configuration clear.

---

## Structural Patterns

### 1. Adapter Pattern

**What it is**: Adapters bridge different serialization protocols (Gob, JSON) and messaging protocols (AMQP, MQTT, Stream).

#### Serialization Adapters

**Location**: `internal/pubsub/consume.go`

**JSON Adapter** (`internal/pubsub/consume.go:77-100`):
```go
// Adapter for JSON serialization
func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange, queueName, key string,
    queueDurability QueueDurability,
    queueType QueueType,
    handler func(T) AckType,
) error {
    return subscribe[T](
        conn, exchange, queueName, key, queueDurability, queueType, handler,
        func(data []byte) (T, error) {  // JSON unmarshalling strategy
            var target T
            err := json.Unmarshal(data, &target)
            return target, err
        },
    )
}
```

**Gob Adapter** (`internal/pubsub/consume.go:102-128`):
```go
// Adapter for Gob serialization
func SubscribeGob[T any](
    conn *amqp.Connection,
    exchange, queueName, key string,
    queueDurability QueueDurability,
    queueType QueueType,
    handler func(T) AckType,
) error {
    return subscribe[T](
        conn, exchange, queueName, key, queueDurability, queueType, handler,
        func(data []byte) (T, error) {  // Gob unmarshalling strategy
            buffer := bytes.NewBuffer(data)
            decoder := gob.NewDecoder(buffer)
            var target T
            err := decoder.Decode(&target)
            return target, err
        },
    )
}
```

**Stream Adapter** (`internal/pubsub/consume.go:48-75`):
```go
// Adapter for RabbitMQ Streams (different protocol than AMQP)
func SubscribeStreamJSON[T any](
    env *stream.Environment,
    streamName string,
    streamOptions *stream.ConsumerOptions,
    handler func(T) AckType,
) (*ha.ReliableConsumer, error) {
    err := DeclareAndBindStream(env, streamName)
    if err != nil && !errors.Is(err, stream.StreamAlreadyExists) {
        return nil, err
    }

    consumer, err := ha.NewReliableConsumer(
        env, streamName, streamOptions,
        func(consumerContext stream.ConsumerContext, message *amqpEncodeStreamMessage.Message) {
            var target T
            err := json.Unmarshal(message.GetData(), &target)
            if err != nil {
                fmt.Printf("could not unmarshal message: %v\n", err)
            }
            handler(target)
        },
    )
    return consumer, err
}
```

**Usage**:
- `cmd/sensor-registry/` uses `SubscribeGob` for sensor registration
- `cmd/sensor-logs-ingester/` uses `SubscribeGob` for log messages
- `cmd/sensor-measurements-ingester/` uses `SubscribeStreamJSON` for high-throughput measurements

**Why Adapter?**:
- ✅ Different protocols for different needs (Gob for efficiency, JSON for compatibility, Streams for throughput)
- ✅ Can swap serialization without changing business logic

#### Protocol Adapters

**MQTT Adapter** (`cmd/sensor-simulation/main.go:229-238`):
```go
// Sensor publishes via MQTT (IoT-friendly protocol)
pubToken := cfg.mqttClient.Publish(
    fmt.Sprintf(routing.KeySensorMeasurements, serialNumber),
    1,     // QoS
    true,  // retain
    payloadBytes,
)
```

**AMQP Adapter** (`cmd/sensor-simulation/main.go:145-153`):
```go
// Sensor publishes registration via AMQP (inter-service protocol)
pubsub.PublishGob(
    publishCh,
    routing.ExchangeTopicIoT,
    fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created",
    routing.Sensor{
        SerialNumber:    serialNumber,
        SampleFrequency: sampleFrequency,
    },
)
```

---

### 2. Facade Pattern

**What it is**: `pubsub` package provides simple interface hiding RabbitMQ complexity.

**Complex Implementation** (`internal/pubsub/consume.go:130-185`):
```go
// Internal complexity hidden from users
func subscribe[T any](
    conn *amqp.Connection,
    exchange, queueName, key string,
    queueDurability QueueDurability,
    queueType QueueType,
    handler func(T) AckType,
    unmarshaller func([]byte) (T, error),
) error {
    // 1. Channel management
    ch, queue, err := DeclareAndBindAMQP(conn, exchange, queueName, key, queueDurability, queueType)
    if err != nil {
        return fmt.Errorf("could not declare and bind queue: %v", err)
    }

    // 2. QoS configuration
    err = ch.Qos(10, 0, false)
    if err != nil {
        return fmt.Errorf("could not set QoS: %v", err)
    }

    // 3. Consumer setup
    msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("could not consume messages: %v", err)
    }

    // 4. Goroutine for message processing
    go func() {
        defer ch.Close()
        for msg := range msgs {
            target, err := unmarshaller(msg.Body)
            if err != nil {
                fmt.Printf("could not unmarshal message: %v\n", err)
                continue
            }
            handler(target)
            msg.Ack(false)
        }
    }()

    return nil
}
```

**Simple Public Interface** (`internal/pubsub/consume.go:77-100, 102-128`):
```go
// Users only see this simple interface
func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange, queueName, key string,
    queueDurability QueueDurability,
    queueType QueueType,
    handler func(T) AckType,
) error {
    return subscribe[T](conn, exchange, queueName, key, queueDurability, queueType, handler, jsonUnmarshaller)
}
```

**Why Facade?**:
- ✅ Hides channel management, QoS, goroutines, acking
- ✅ Users only need to provide handler function
- ✅ Consistent error handling

---

### 3. Strategy Pattern

**What it is**: Interchangeable serialization algorithms (Gob vs JSON) passed as strategies.

**Strategy Interface** (`internal/pubsub/consume.go:138`):
```go
unmarshaller func([]byte) (T, error)  // Strategy function type
```

**Strategy 1: JSON** (`internal/pubsub/consume.go:94-98`):
```go
func(data []byte) (T, error) {
    var target T
    err := json.Unmarshal(data, &target)
    return target, err
}
```

**Strategy 2: Gob** (`internal/pubsub/consume.go:119-125`):
```go
func(data []byte) (T, error) {
    buffer := bytes.NewBuffer(data)
    decoder := gob.NewDecoder(buffer)
    var target T
    err := decoder.Decode(&target)
    return target, err
}
```

**Context** (`internal/pubsub/consume.go:130-185`):
```go
func subscribe[T any](
    // ... params ...
    unmarshaller func([]byte) (T, error),  // Strategy injected
) error {
    // ... setup ...
    go func() {
        for msg := range msgs {
            target, err := unmarshaller(msg.Body)  // Strategy used here
            if err != nil {
                fmt.Printf("could not unmarshal message: %v\n", err)
                continue
            }
            handler(target)
            msg.Ack(false)
        }
    }()
    return nil
}
```

**Why Strategy?**:
- ✅ Can add new serialization formats (e.g., Protobuf) without modifying `subscribe()`
- ✅ Choose format based on use case (Gob for performance, JSON for compatibility)

---

## Behavioral Patterns

### 1. Observer Pattern (Pub/Sub)

**What it is**: Services observe events via message broker without knowing who publishes.

**Subject (Publisher)** (`cmd/sensor-simulation/main.go:145-153`):
```go
// Publisher doesn't know WHO is observing
pubsub.PublishGob(
    publishCh,
    routing.ExchangeTopicIoT,
    fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created",
    routing.Sensor{
        SerialNumber:    serialNumber,
        SampleFrequency: sampleFrequency,
    },
)
```

**Observer 1: sensor-registry** (observes sensor registration):
```go
// Observer 1 subscribes to sensor registration events
err = pubsub.SubscribeGob(
    conn,
    routing.ExchangeTopicIoT,
    routing.QueueSensorRegistry,
    fmt.Sprintf(routing.KeySensorRegistryFormat, "*")+"."+"#",
    pubsub.QueueDurable,
    pubsub.QueueClassic,
    handlerSensorRegistry(ctx, db),  // Observer's reaction
)
```

**Multiple Observers Possible**:
- `sensor-registry` observes registration events → writes to DB
- Future: `sensor-notifier` could observe same events → send email
- Future: `sensor-analytics` could observe same events → update dashboard

**Why Observer?**:
- ✅ Loose coupling: Publisher doesn't know about observers
- ✅ Open/Closed: Add new observers without modifying publisher
- ✅ Dynamic: Observers can subscribe/unsubscribe at runtime

---

### 2. Command Pattern

**What it is**: Encapsulates requests as message objects with parameters.

**Command Message DTO** (`internal/routing/models.go:21-26`):
```go
// Command encapsulated as data
type SensorCommandMessage struct {
    SerialNumber string
    Timestamp    time.Time
    Command      string                 // "sleep", "awake", "changeSampleFrequency"
    Params       map[string]interface{} // {"sampleFrequency": 1000}
}
```

**Command Constants** (`internal/sensorlogic/commands.go`):
```go
const (
    CommandLogin                 = "login"
    CommandLogout                = "logout"
    CommandSleep                 = "sleep"
    CommandAwake                 = "awake"
    CommandChangeSampleFrequency = "changeSampleFrequency"
    CommandDelete                = "delete"
)
```

**Command Invoker** (`cmd/sensor-simulation/handlers.go` - handler pattern):
```go
func handlerCommand(sensorState *sensorlogic.SensorState) func(cm routing.SensorCommandMessage) pubsub.AckType {
    return func(cm routing.SensorCommandMessage) pubsub.AckType {
        // Dispatches command to appropriate handler
        switch cm.Command {
        case "sleep":
            sensorState.HandleSleep()
        case "awake":
            sensorState.HandleAwake()
        case "changeSampleFrequency":
            sensorState.HandleChangeSampleFrequency(cm.Params)
        default:
            sensorState.LogsWarning <- "not a valid command"
            return pubsub.NackDiscard
        }
        return pubsub.Ack
    }
}
```

**Command Execution** (`internal/sensorlogic/sensor.go` - assuming methods exist):
```go
// Concrete command implementations
func (sensorState *SensorState) HandleSleep() {
    if sensorState.IsSleep {
        sensorState.LogsInfo <- "sensor is already in a sleep state"
        return
    }
    sensorState.IsSleep = true
    sensorState.IsSleepChan <- true
}

func (sensorState *SensorState) HandleAwake() {
    if sensorState.IsSleep {
        sensorState.IsSleep = false
        sensorState.IsSleepChan <- false
        sensorState.LogsInfo <- "sensor is awake from sleep"
        return
    }
    sensorState.LogsInfo <- "sensor is already in an awake state"
}
```

**Why Command?**:
- ✅ Commands are serializable (can be queued, logged, replayed)
- ✅ Decouples sender (iot-api) from receiver (sensor-simulation)
- ✅ Easy to add new commands without changing command execution infrastructure

---

### 3. State Pattern

**What it is**: Object changes behavior based on internal state.

**State Object** (`internal/sensorlogic/sensor.go:12-21`):
```go
type SensorState struct {
    Sensor                    Sensor       // Sensor entity
    LogsInfo                  chan string  // Log channels
    LogsWarning               chan string
    LogsError                 chan string
    SampleFrequency           float64      // Current frequency
    SampleFrequencyChangeChan chan float64 // Frequency change channel
    IsSleep                   bool         // STATE: Is sensor asleep?
    IsSleepChan               chan bool    // STATE CHANGE channel
}
```

**State-Dependent Behavior** (`cmd/sensor-simulation/main.go:242-249`):
```go
// Behavior changes based on IsSleep state
case isSleep := <-sensorState.IsSleepChan:
    if isSleep {
        ticker.Stop()         // Stop collecting measurements (SLEEP state behavior)
        batchTimer.Stop()
    } else {
        ticker = time.NewTicker(time.Second / time.Duration(sensorState.SampleFrequency))
        batchTimer = time.NewTicker(batchTime)  // Resume measurements (AWAKE state behavior)
    }
```

**State Transitions** (`internal/sensorlogic/sensor.go` - assuming methods):
```go
// Transition from AWAKE → SLEEP
func (sensorState *SensorState) HandleSleep() {
    if sensorState.IsSleep {
        sensorState.LogsInfo <- "sensor is already in a sleep state"
        return
    }
    sensorState.IsSleep = true  // STATE CHANGE
    sensorState.IsSleepChan <- true
}

// Transition from SLEEP → AWAKE
func (sensorState *SensorState) HandleAwake() {
    if sensorState.IsSleep {
        sensorState.IsSleep = false  // STATE CHANGE
        sensorState.IsSleepChan <- false
        sensorState.LogsInfo <- "sensor is awake from sleep"
        return
    }
    sensorState.LogsInfo <- "sensor is already in an awake state"
}
```

**States**:
- **AWAKE**: Ticker running, measurements collected, batch timer active
- **SLEEP**: Tickers stopped, no measurements, low power mode

**Why State Pattern?**:
- ✅ Encapsulates state-dependent behavior
- ✅ Makes state transitions explicit
- ✅ Easy to add new states (e.g., CALIBRATING, ERROR)

---

### 4. Handler/Callback Pattern

**What it is**: Functions returned as handlers for processing messages.

**Handler Factory** (`cmd/sensor-registry/handlers.go` - assuming similar structure):
```go
// Factory returns a handler with closure over db
func handlerSensorRegistry(ctx context.Context, db *storage.DB) func(dto routing.Sensor) pubsub.AckType {
    return func(dto routing.Sensor) pubsub.AckType {
        // DTO → DB Record transformation
        record := storage.SensorRecord{
            SerialNumber:    dto.SerialNumber,
            SampleFrequency: dto.SampleFrequency,
        }

        err := db.WriteSensor(ctx, record)
        if err != nil {
            return pubsub.NackRequeue  // Retry on error
        }
        return pubsub.Ack
    }
}
```

**Handler Registration** (`cmd/sensor-registry/main.go` - assuming):
```go
// Handler passed to subscription
err = pubsub.SubscribeGob(
    conn,
    routing.ExchangeTopicIoT,
    routing.QueueSensorRegistry,
    fmt.Sprintf(routing.KeySensorRegistryFormat, "*")+"."+"#",
    pubsub.QueueDurable,
    pubsub.QueueClassic,
    handlerSensorRegistry(ctx, db),  // Handler with closed-over dependencies
)
```

**Why Handler/Callback?**:
- ✅ Handlers have access to dependencies via closure (db, ctx)
- ✅ Clean separation: subscription logic vs business logic
- ✅ Testable: Can test handler independently

---

## Concurrency Patterns

### 1. Worker Pool Pattern

**What it is**: Goroutines process messages from channels.

**Implementation** (`cmd/sensor-simulation/main.go:103-129`):
```go
// Worker goroutine processes log messages from 3 channels
go func() {
    for {
        select {
        case infoMsg := <-sensorState.LogsInfo:
            publishSensorLog(publishCh, routing.SensorLog{
                SerialNumber: serialNumber,
                Timestamp:    time.Now(),
                Level:        "INFO",
                Message:      infoMsg,
            })
        case warningMsg := <-sensorState.LogsWarning:
            publishSensorLog(publishCh, routing.SensorLog{
                SerialNumber: serialNumber,
                Timestamp:    time.Now(),
                Level:        "WARNING",
                Message:      warningMsg,
            })
        case errMsg := <-sensorState.LogsError:
            publishSensorLog(publishCh, routing.SensorLog{
                SerialNumber: serialNumber,
                Timestamp:    time.Now(),
                Level:        "ERROR",
                Message:      errMsg,
            })
        }
    }
}()
```

**Work Submission** (`cmd/sensor-simulation/main.go:131-141`):
```go
// Submitting work to the worker via channels
sensorState.LogsInfo <- "System powering on..."
time.Sleep(100 * time.Millisecond)
sensorState.LogsInfo <- "Bootloader version: v1.0.0"
time.Sleep(200 * time.Millisecond)
sensorState.LogsInfo <- "Loading configuration..."
sensorState.LogsInfo <- "Configuration loaded successfully"
time.Sleep(100 * time.Millisecond)
sensorState.LogsInfo <- "Performing sensor self-test"
time.Sleep(500 * time.Millisecond)
sensorState.LogsInfo <- "Self-test result: PASSED"
```

**Why Worker Pool?**:
- ✅ Non-blocking: Logging doesn't block main sensor loop
- ✅ Sequential processing: Logs published in order
- ✅ Buffered channels (size 1) prevent blocking on single log

---

### 2. Producer-Consumer Pattern

**What it is**: Sensors produce data, ingesters consume, with RabbitMQ as buffer.

**Producer** (`cmd/sensor-simulation/main.go:217-240`):
```go
// PRODUCER: Batch measurements and publish
case <-batchTimer.C:
    if len(measurements) == 0 {
        continue // nothing to send...
    }
    payloadBytes, err := json.Marshal(measurements)
    if err != nil {
        log.Printf("Failed to marshal measurements: %v", err)
        return
    }

    // PUBLISH to queue
    pubToken := cfg.mqttClient.Publish(
        fmt.Sprintf(routing.KeySensorMeasurements, serialNumber),
        1,
        true,
        payloadBytes,
    )
    pubToken.Wait()
    if pubToken.Error() != nil {
        log.Printf("Publish error: %v", pubToken.Error())
    }

    measurements = measurements[:0]  // Clear batch
```

**Buffer** (`internal/routing/routing.go:14`):
```go
// RabbitMQ Stream acts as persistent buffer between producer and consumer
const QueueSensorMeasurements = "sensor.all.measurements.db_writer"
```

**Consumer** (`cmd/sensor-measurements-ingester/main.go` - assuming pattern):
```go
// CONSUMER: Subscribe to stream and persist to DB
consumer, err := pubsub.SubscribeStreamJSON(
    env,
    routing.QueueSensorMeasurements,
    stream.NewConsumerOptions().
        SetOffset(stream.OffsetSpecification{}.First()).
        SetConsumerName(routing.StreamConsumerName),
    handlerMeasurements(db, ctx),  // Handler persists to PostgreSQL
)
```

**Why Producer-Consumer?**:
- ✅ Decoupling: Producers and consumers run at different speeds
- ✅ Buffering: Queue absorbs bursts of measurements
- ✅ Durability: Messages persist if consumer goes down

---

### 3. Pipeline Pattern

**What it is**: Data flows through multiple transformation stages.

**Pipeline Stages**:

```
Stage 1: COLLECTION (sensor-simulation)
   ↓ Collect measurements from simulated signal
Stage 2: BATCHING (sensor-simulation)
   ↓ Batch N measurements together
Stage 3: PUBLISHING (sensor-simulation → RabbitMQ)
   ↓ Publish batch via MQTT
Stage 4: CONSUMPTION (sensor-measurements-ingester)
   ↓ Consume from RabbitMQ Stream
Stage 5: TRANSFORMATION (sensor-measurements-ingester)
   ↓ Map DTO → DB Record, lookup sensor IDs
Stage 6: PERSISTENCE (PostgreSQL/TimescaleDB)
   ↓ Batch insert into database
```

**Stage 1: Collection** (`cmd/sensor-simulation/main.go:200-215`):
```go
case <-ticker.C:
    accX, timestamp := func() (float64, time.Duration) {
        timestamp := time.Since(startTime)
        elapsedSec := timestamp.Seconds()

        value := sensorlogic.SimulateSignal(sineWaves, elapsedSec)

        return value, timestamp
    }()

    // Collect into batch
    measurements = append(measurements, routing.SensorMeasurement{
        SerialNumber: serialNumber,
        Timestamp:    startTime.Add(timestamp),
        Value:        accX,
    })
```

**Stage 2: Batching** (`cmd/sensor-simulation/main.go:217-240`):
```go
case <-batchTimer.C:
    if len(measurements) == 0 {
        continue
    }
    payloadBytes, err := json.Marshal(measurements)  // Marshal batch
    // ... publish ...
    measurements = measurements[:0]  // Clear batch
```

**Stage 3: Publishing** (`cmd/sensor-simulation/main.go:229-238`):
```go
pubToken := cfg.mqttClient.Publish(
    fmt.Sprintf(routing.KeySensorMeasurements, serialNumber),
    1,
    true,
    payloadBytes,
)
```

**Stage 4: Consumption** (automatic via `pubsub.SubscribeStreamJSON`)

**Stage 5: Transformation** (`internal/sensorlogic/sensormeasurements.go:111-140`):
```go
func HandleMeasurements(ctx context.Context, db *storage.DB, dtos []routing.SensorMeasurement) error {
    // Fetch sensor ID mapping
    sensorMap, err := db.GetSensorIDBySerialNumberMap(ctx)
    if err != nil {
        return fmt.Errorf("failed to fetch sensor IDs: %v", err)
    }

    // Transform DTO → DB Record
    records := make([]storage.SensorMeasurementRecord, len(dtos))
    for i, dto := range dtos {
        sensorID, exists := sensorMap[dto.SerialNumber]
        if !exists {
            return fmt.Errorf("sensor serial number not found: %s", dto.SerialNumber)
        }
        records[i] = storage.SensorMeasurementRecord{
            Timestamp:   dto.Timestamp,
            SensorID:    sensorID,
            Measurement: dto.Value,
        }
    }

    // Proceed to next stage
    return db.BatchArrayWriteMeasurement(ctx, records)
}
```

**Stage 6: Persistence** (`internal/storage/measurements.go` - assuming batch write):
```go
func (DB *DB) BatchArrayWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error {
    // ... batch insert logic ...
}
```

**Why Pipeline?**:
- ✅ Separation of concerns: Each stage has single responsibility
- ✅ Testable: Can test each stage independently
- ✅ Performance: Batching reduces DB roundtrips

---

### 4. Single Active Consumer Pattern

**What it is**: Only one consumer processes stream messages at a time for consistent ordering.

**Implementation** (`cmd/sensor-measurements-ingester/main.go` - assuming pattern):
```go
// Callback invoked when consumer becomes active/inactive
singleActiveConsumerUpdate := func(streamName string, isActive bool) stream.OffsetSpecification {
    fmt.Printf("[%s] - Consumer promoted for: %s. Active status: %t\n",
        time.Now().Format(time.TimeOnly), streamName, isActive)

    // Query last processed offset
    offset, err := env.QueryOffset(routing.StreamConsumerName, routing.QueueSensorMeasurements)
    if err != nil {
        return stream.OffsetSpecification{}.First()
    }

    // Resume from next offset
    return stream.OffsetSpecification{}.Offset(offset + 1)
}

// Register consumer with single-active-consumer constraint
consumer, err := pubsub.SubscribeStreamJSON(
    env,
    routing.QueueSensorMeasurements,
    stream.NewConsumerOptions().
        SetOffset(stream.OffsetSpecification{}.First()).
        SetConsumerName(routing.StreamConsumerName).
        SetSingleActiveConsumer(stream.NewSingleActiveConsumer(singleActiveConsumerUpdate)),
    handlerMeasurements(db, ctx),
)
```

**Why Single Active Consumer?**:
- ✅ Ordering: Measurements processed in order (critical for time-series)
- ✅ High availability: If active consumer dies, another becomes active
- ⚠️ Trade-off: Lower throughput than parallel consumers

---

### 5. Refresh Loop Pattern (Cache Invalidation)

**What it is**: Goroutine periodically refreshes cached data.

**Cache Structure** (`internal/sensorlogic/sensormeasurements.go:14-19`):
```go
type SensorCache struct {
    mu          sync.RWMutex         // Read-Write lock for concurrent access
    mapping     map[string]int       // Serial Number → Sensor ID
    lastRefresh time.Time            // Last refresh timestamp
    db          *storage.DB          // DB connection for refresh
}
```

**Factory with Initial Load** (`internal/sensorlogic/sensormeasurements.go:21-34`):
```go
func NewSensorCache(ctx context.Context, db *storage.DB) (*SensorCache, error) {
    cache := &SensorCache{
        db:      db,
        mapping: make(map[string]int),
    }

    // Initial load
    err := cache.refresh(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed initial sensor cache load: %v", err)
    }

    return cache, nil
}
```

**Refresh Logic** (`internal/sensorlogic/sensormeasurements.go:36-49`):
```go
func (sc *SensorCache) refresh(ctx context.Context) error {
    // Fetch latest data from DB
    sensorMap, err := sc.db.GetSensorIDBySerialNumberMap(ctx)
    if err != nil {
        return fmt.Errorf("failed to fetch sensor IDs: %v", err)
    }

    // Update cache atomically
    sc.mu.Lock()
    sc.mapping = sensorMap
    sc.lastRefresh = time.Now()
    sc.mu.Unlock()

    fmt.Printf("[%s] Sensor cache refreshed: %d sensors loaded\n", time.Now().Format(time.RFC3339), len(sensorMap))

    return nil
}
```

**Refresh Loop** (`internal/sensorlogic/sensormeasurements.go:51-67`):
```go
func (sc *SensorCache) StartRefreshLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)  // Refresh every 30 seconds
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            err := sc.refresh(ctx)
            if err != nil {
                fmt.Printf("ERROR: sensor cache refresh failed: %v\n", err)
            }
        case <-ctx.Done():
            fmt.Println("Sensor cache refresh loop stopped")
            return
        }
    }
}
```

**Thread-Safe Read** (`internal/sensorlogic/sensormeasurements.go:69-74`):
```go
func (sc *SensorCache) Get(serialNumber string) (int, bool) {
    sc.mu.RLock()         // Read lock (allows concurrent reads)
    defer sc.mu.RUnlock()
    id, exists := sc.mapping[serialNumber]
    return id, exists
}
```

**Usage**:
```go
// Startup
cache, err := sensorlogic.NewSensorCache(ctx, db)
go cache.StartRefreshLoop(ctx)  // Start background refresh

// During message processing
sensorID, exists := cache.Get(serialNumber)  // Fast, no DB query
```

**Why Refresh Loop?**:
- ✅ Performance: Avoids DB query on every message (cache hit rate >> 99%)
- ✅ Eventual consistency: New sensors available within 30 seconds
- ✅ Thread-safe: RWMutex allows concurrent reads, exclusive writes

---

## Data Patterns

### 1. DTO (Data Transfer Object) Pattern

**What it is**: DTOs define message contracts between services.

**Location**: `internal/routing/models.go`

**DTO 1: Sensor Registration** (`internal/routing/models.go:8-11`):
```go
// DTO for sensor enrollment
type Sensor struct {
    SerialNumber    string
    SampleFrequency float64
}
```

**DTO 2: Measurements** (`internal/routing/models.go:14-18`):
```go
// DTO for sensor measurements
type SensorMeasurement struct {
    SerialNumber string
    Timestamp    time.Time
    Value        float64
}
```

**DTO 3: Commands** (`internal/routing/models.go:21-26`):
```go
// DTO for sensor commands (sleep, awake, changeSampleFrequency)
type SensorCommandMessage struct {
    SerialNumber string
    Timestamp    time.Time
    Command      string                 // e.g., "sleep", "awake"
    Params       map[string]interface{} // e.g., {"sampleFrequency": 1000}
}
```

**DTO 4: Logs** (`internal/routing/models.go:29-34`):
```go
// DTO for sensor logs
type SensorLog struct {
    SerialNumber string
    Timestamp    time.Time
    Level        string   // "INFO", "WARNING", "ERROR"
    Message      string
}
```

**DTO → DB Record Mapping** (`internal/sensorlogic/sensormeasurements.go:119-132`):
```go
// Transform DTO (wire format) to DB Record (storage format)
for i, dto := range dtos {
    sensorID, exists := sensorMap[dto.SerialNumber]
    if !exists {
        return fmt.Errorf("sensor serial number not found: %s", dto.SerialNumber)
    }

    // Map DTO fields → DB Record fields
    records[i] = storage.SensorMeasurementRecord{
        Timestamp:   dto.Timestamp,
        SensorID:    sensorID,           // String → Int (foreign key)
        Measurement: dto.Value,
    }
}
```

**Why DTO?**:
- ✅ Decoupling: Wire format ≠ storage format
- ✅ Versioning: Can change DB schema without breaking message contracts
- ✅ Validation: DTOs can have different validation rules than DB records

---

### 2. Repository Pattern

**What it is**: Database access abstraction layer.

**Repository Interface** (`internal/storage/db.go:16-39`):
```go
// Repository wraps database connection
type DB struct {
    pool *pgxpool.Pool  // Connection pool hidden from users
}

func NewDBPool(connString string) (*DB, error) {
    ctx := context.Background()
    dbpool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %v", err)
    }
    return &DB{pool: dbpool}, nil
}

func (db *DB) Close() {
    db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
    return db.pool.Ping(ctx)
}
```

**Repository Methods - Sensors** (`internal/storage/sensors.go`):

**Find by Serial Number** (`internal/storage/sensors.go:12-26`):
```go
func (db *DB) GetSensorIDBySerialNumber(ctx context.Context, serialNumber string) (sensorID int, err error) {
    queryGetSensor := `
        SELECT id
        FROM sensor
        WHERE serial_number = ($1)
    ;`

    err = db.pool.QueryRow(ctx, queryGetSensor, serialNumber).Scan(&sensorID)
    if err != nil {
        return 0, fmt.Errorf("unable to query sensor ID: %v", err)
    }

    return sensorID, nil
}
```

**Find All as Map** (`internal/storage/sensors.go:77-104`):
```go
func (db *DB) GetSensorIDBySerialNumberMap(ctx context.Context) (map[string]int, error) {
    sensorMap := make(map[string]int)

    query := `SELECT serial_number, id FROM sensor`

    rows, err := db.pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch sensor IDs: %v", err)
    }
    defer rows.Close()

    for rows.Next() {
        var serialNumber string
        var sensorID int

        if err := rows.Scan(&serialNumber, &sensorID); err != nil {
            return nil, fmt.Errorf("failed to scan row: %v", err)
        }

        sensorMap[serialNumber] = sensorID
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %v", err)
    }

    return sensorMap, nil
}
```

**Create** (`internal/storage/sensors.go:106-138`):
```go
func (db *DB) WriteSensor(ctx context.Context, sr SensorRecord) error {
    // Check if exists
    queryCheckIfExists := `SELECT EXISTS (
        SELECT 1 FROM sensor WHERE serial_number = ($1)
    );`

    var rowExists bool
    err := db.pool.QueryRow(ctx, queryCheckIfExists, sr.SerialNumber).Scan(&rowExists)
    if err != nil {
        log.Fatal(err)
    }

    if rowExists {
        fmt.Printf("Entry for sensor `%s` already exists. Skipping...\n", sr.SerialNumber)
        return nil
    }

    // Insert
    queryInsertMetadata := `INSERT INTO sensor (serial_number, sample_frequency) VALUES ($1, $2);`

    _, err = db.pool.Exec(ctx, queryInsertMetadata, sr.SerialNumber, sr.SampleFrequency)
    if err != nil {
        return fmt.Errorf("unable to insert sensor metadata into database: %v", err)
    }
    fmt.Printf("Inserted sensor (%s) into `sensor` table\n", sr.SerialNumber)

    return nil
}
```

**Delete** (`internal/storage/sensors.go:140-170`):
```go
func (db *DB) DeleteSensor(ctx context.Context, serialNumber string) error {
    // Check if exists
    queryCheckIfExists := `SELECT EXISTS (
        SELECT 1 FROM sensor WHERE serial_number = ($1)
    );`

    var rowExists bool
    err := db.pool.QueryRow(ctx, queryCheckIfExists, serialNumber).Scan(&rowExists)
    if err != nil {
        log.Fatal(err)
    }

    if !rowExists {
        fmt.Printf("Entry for sensor `%s` does not exist. Skipping...\n", serialNumber)
        return nil
    }

    // Delete
    queryDeleteMetadata := `DELETE FROM sensor WHERE serial_number = ($1);`

    _, err = db.pool.Exec(ctx, queryDeleteMetadata, serialNumber)
    if err != nil {
        return fmt.Errorf("unable to delete sensor metadata from database: %v", err)
    }
    fmt.Printf("Deleted sensor (%s) from `sensor` table (and all its measurements)\n", serialNumber)
    return nil
}
```

**Repository Methods - Measurements** (`internal/storage/measurements.go`):

**Batch Write (multiple strategies available)**:
```go
// Strategy 1: Value strings (simple)
func (DB *DB) BatchWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error

// Strategy 2: Array unnest (PostgreSQL-specific, faster)
func (DB *DB) BatchArrayWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error

// Strategy 3: COPY protocol (fastest)
func (DB *DB) CopyWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error
```

**Why Repository?**:
- ✅ Abstraction: Business logic doesn't know SQL
- ✅ Testability: Can mock `DB` interface for testing
- ✅ Centralization: All queries in one place (easier to optimize)

---

## HTTP/Response Patterns

### 1. Response Helper Pattern

**What it is**: Centralized JSON and error response formatting.

**Location**: `cmd/iot-api/json.go`

**Error Response Helper** (`cmd/iot-api/json.go` - assuming pattern):
```go
func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
    if err != nil {
        log.Println(err)  // Log internal error
    }
    if code > 499 {
        log.Println("Responding with 5XX error:", msg)
    }

    type errorResponse struct {
        Error string `json:"error"`
    }

    respondWithJSON(w, code, errorResponse{Error: msg})
}
```

**JSON Response Helper** (`cmd/iot-api/json.go` - assuming pattern):
```go
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)

    dat, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshalling JSON: %s", err)
        w.WriteHeader(500)
        return
    }

    w.Write(dat)
}
```

**Usage in Handlers** (`cmd/iot-api/handler_sensors_awake.go` - assuming pattern):
```go
func (cfg *apiConfig) handlerSensorsAwake(w http.ResponseWriter, req *http.Request) {
    sensorSerialNumber := req.PathValue("sensorSerialNumber")

    publishCh, err := cfg.rabbitConn.Channel()
    if err != nil {
        respondWithError(w, 500, "could not create channel to publish sensor's awake command", err)
        return
    }
    defer publishCh.Close()

    // ... business logic ...

    respondWithJSON(w, 200, map[string]string{"status": "awake command sent"})
}
```

**Why Response Helper?**:
- ✅ Consistency: All responses have same format
- ✅ DRY: No repeated header/marshal code
- ✅ Logging: Errors logged in one place

---

## Configuration Patterns

### 1. Configuration as Constants

**What it is**: Centralized routing configuration prevents typos and inconsistencies.

**Location**: `internal/routing/routing.go`

**Connection Strings** (`internal/routing/routing.go:4-5` - assuming):
```go
var (
    RabbitConnString     = os.Getenv("RABBITMQ_CONN_STRING")
    RabbitMQTTConnString = os.Getenv("RABBITMQ_MQTT_CONN_STRING")
)
```

**Exchange Configuration** (`internal/routing/routing.go:7-9`):
```go
const (
    ExchangeTopicIoT = "iot"  // Single exchange for all IoT messages
)
```

**Queue Naming Convention** (`internal/routing/routing.go:12-20`):
```go
// Queue naming pattern: entity.id.consumer.type
const (
    QueueSensorMeasurements   = "sensor.all.measurements.db_writer"
    QueueSensorCommandsFormat = "sensor.%s.commands"        // %s = serial number
    QueueSensorRegistry       = "sensor.all.registry.created"
    QueueSensorLogs           = "sensor.all.logs"
)
```

**Routing Key Patterns** (`internal/routing/routing.go:23-31`):
```go
// Routing keys for topic exchange
const (
    KeySensorMeasurements   = "sensor.%s.measurements"  // %s = serial number
    KeySensorCommandsFormat = "sensor.%s.commands"
    KeySensorRegistryFormat = "sensor.%s.registry"
    KeySensorLogsFormat     = "sensor.%s.logs"
)
```

**Stream Configuration** (`internal/routing/routing.go` - assuming):
```go
const (
    StreamConsumerName = "sensor-measurements-ingester"
)
```

**Usage**:
```go
// Instead of typo-prone string literals:
// ❌ pubsub.PublishGob(ch, "iot", "sensor.ABC123.registry.created", sensor)

// Use constants:
// ✅ pubsub.PublishGob(ch, routing.ExchangeTopicIoT,
//                      fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created",
//                      sensor)
```

**Why Constants?**:
- ✅ Single source of truth for routing topology
- ✅ Compiler catches typos
- ✅ Easy refactoring (rename in one place)

---

## Design Pattern Quick Reference

| Pattern | Location | Line Reference |
|---------|----------|----------------|
| **Hexagonal Architecture** | `internal/` structure | Entire `internal/` directory |
| **Event-Driven Pub/Sub** | `internal/pubsub/` | `consume.go`, `publish.go` |
| **Microservices** | `cmd/` | 6 services |
| **Factory - DB** | `internal/storage/db.go` | Line 20-31 |
| **Factory - API Config** | `cmd/iot-api/main.go` | Line 21-36 |
| **Factory - Sensor State** | `internal/sensorlogic/sensor.go` | Line 23-36 |
| **Factory - Cache** | `internal/sensorlogic/sensormeasurements.go` | Line 21-34 |
| **Singleton - Connection Pool** | `internal/storage/db.go` | Line 16-18 |
| **Builder - MQTT** | `cmd/sensor-simulation/main.go` | Line 32-42 |
| **Adapter - JSON** | `internal/pubsub/consume.go` | Line 77-100 |
| **Adapter - Gob** | `internal/pubsub/consume.go` | Line 102-128 |
| **Adapter - Stream** | `internal/pubsub/consume.go` | Line 48-75 |
| **Facade - Pub/Sub** | `internal/pubsub/consume.go` | Line 130-185 |
| **Strategy - Serialization** | `internal/pubsub/consume.go` | Line 138 (func param) |
| **Observer - Pub/Sub** | All consumers | Various |
| **Command** | `internal/routing/models.go` | Line 21-26 |
| **State** | `internal/sensorlogic/sensor.go` | Line 12-21 |
| **Handler/Callback** | `cmd/sensor-registry/handlers.go` | Handler functions |
| **Worker Pool** | `cmd/sensor-simulation/main.go` | Line 103-129 |
| **Producer-Consumer** | `cmd/sensor-simulation/` + `cmd/sensor-measurements-ingester/` | Both services |
| **Pipeline** | Data flow across services | Multi-service |
| **Single Active Consumer** | `cmd/sensor-measurements-ingester/` | Stream consumer setup |
| **Refresh Loop** | `internal/sensorlogic/sensormeasurements.go` | Line 51-67 |
| **DTO** | `internal/routing/models.go` | Line 1-34 |
| **Repository** | `internal/storage/` | `sensors.go`, `measurements.go` |
| **Response Helper** | `cmd/iot-api/json.go` | Entire file |
| **Config Constants** | `internal/routing/routing.go` | Entire file |

---

## Adding New Features - Pattern Checklist

When adding new functionality, consult this checklist to maintain consistency:

### Adding a New Message Type

1. **DTO**: Define in `internal/routing/models.go`
2. **Routing Key**: Add constant to `internal/routing/routing.go`
3. **Queue Name**: Add constant to `internal/routing/routing.go`
4. **Publisher**: Use `pubsub.PublishGob()` or `pubsub.PublishJSON()`
5. **Subscriber**: Use `pubsub.SubscribeGob()` or `pubsub.SubscribeJSON()`
6. **Handler**: Create handler factory returning `func(DTO) pubsub.AckType`

### Adding a New Service

1. **Directory**: Create `cmd/my-service/`
2. **Main**: Create `main.go` with dependency injection (Factory pattern)
3. **Handlers**: Create `handlers.go` with handler factories
4. **Config**: Use `routing` constants for queues/exchanges
5. **Graceful Shutdown**: Add signal handling (`syscall.SIGINT`, `syscall.SIGTERM`)

### Adding a New Repository Method

1. **Location**: `internal/storage/`
2. **Receiver**: Add method to `DB` struct
3. **Context**: Accept `context.Context` as first parameter
4. **Error Handling**: Return descriptive errors with `fmt.Errorf()`
5. **Naming**: Follow pattern `Get*`, `Write*`, `Delete*`

### Adding a New Concurrency Pattern

1. **Channels**: Define in state struct (buffered if needed)
2. **Goroutine**: Launch with clear lifecycle (defer cleanup)
3. **Synchronization**: Use `sync.RWMutex` for shared state
4. **Shutdown**: Listen to `ctx.Done()` for graceful shutdown

---

## Common Pitfalls to Avoid

### Anti-Pattern: Skipping Hexagonal Architecture

❌ **Bad**: Business logic directly imports `pgx` or `amqp091-go`
```go
// internal/sensorlogic/bad.go
import "github.com/jackc/pgx/v5/pgxpool"

func HandleMeasurements(connString string, dtos []routing.SensorMeasurement) error {
    pool, _ := pgxpool.New(ctx, connString)  // ❌ Business logic knows about pgx
    // ...
}
```

✅ **Good**: Business logic depends on repository interface
```go
// internal/sensorlogic/sensormeasurements.go
func HandleMeasurements(ctx context.Context, db *storage.DB, dtos []routing.SensorMeasurement) error {
    sensorMap, err := db.GetSensorIDBySerialNumberMap(ctx)  // ✅ Uses interface
    // ...
}
```

### Anti-Pattern: Magic Strings

❌ **Bad**: Hardcoded routing keys
```go
pubsub.PublishGob(ch, "iot", "sensor.ABC123.measurements", data)  // ❌ Typo-prone
```

✅ **Good**: Use constants from `routing` package
```go
pubsub.PublishGob(ch, routing.ExchangeTopicIoT,
                  fmt.Sprintf(routing.KeySensorMeasurements, serialNumber),
                  data)  // ✅ Type-safe
```

### Anti-Pattern: Ignoring Cache in High-Throughput Path

❌ **Bad**: DB query per message
```go
func handlerMeasurements(db *storage.DB, ctx context.Context) func([]routing.SensorMeasurement) pubsub.AckType {
    return func(measurements []routing.SensorMeasurement) pubsub.AckType {
        for _, m := range measurements {
            sensorID, _ := db.GetSensorIDBySerialNumber(ctx, m.SerialNumber)  // ❌ DB query per measurement!
            // ...
        }
    }
}
```

✅ **Good**: Use cache
```go
func handlerMeasurements(cache *SensorCache, db *storage.DB, ctx context.Context) func([]routing.SensorMeasurement) pubsub.AckType {
    return func(measurements []routing.SensorMeasurement) pubsub.AckType {
        sensorMap := cache.GetAll()  // ✅ Single cache read
        for _, m := range measurements {
            sensorID := sensorMap[m.SerialNumber]
            // ...
        }
    }
}
```

---

## Conclusion

This document captures the **design wisdom** embedded in the sensor-data-streaming-pubsub repository. When making changes:

1. ✅ Follow existing patterns for consistency
2. ✅ Consult this document to understand "why" patterns exist
3. ✅ Update this document when introducing new patterns
4. ✅ Use file references to quickly locate implementations

**Remember**: Patterns exist to solve problems. Understand the problem before applying the pattern.

---

**Document Metadata**:
- **Author**: Repository Analysis
- **Last Updated**: 2025-11-17
- **Repository**: sensor-data-streaming-pubsub
- **Purpose**: Architectural reference for maintaining consistency
