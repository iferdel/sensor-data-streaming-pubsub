# MQTT Client ID Duplicate Issue

## TL;DR

**Problem:** Only one sensor publishing data, others keep disconnecting.

**Cause:** Multiple sensors using the same MQTT client ID ("publisher").

**Fix:** Use unique client IDs (sensor serial numbers).

**Verify fix:** `docker logs iot-rabbitmq | grep "duplicate"` should show no recent warnings.

---

## Problem Summary

When running multiple sensor simulations, only one sensor was sending measurements to the database while the other sensor appeared to connect but wasn't publishing data. This was caused by both sensors using the same MQTT client ID.

## What is an MQTT Client ID?

An MQTT client ID is a **unique identifier** that each MQTT client must provide when connecting to an MQTT broker (in our case, RabbitMQ with MQTT plugin enabled). The client ID is used by the broker to:

1. Track which client is connected
2. Maintain session state for the client
3. Handle Quality of Service (QoS) message delivery
4. Manage subscriptions and unacknowledged messages

## Why Can't Two Clients Share the Same Client ID?

According to the MQTT specification (version 3.1.1 and 5.0), **a client ID must be unique** within the broker. When a new client connects with the same client ID as an already-connected client:

1. The broker **disconnects the existing client** with that ID
2. The new client takes over the connection
3. Any in-flight messages or subscriptions from the old client are terminated

### What Was Happening in Our System

```
Timeline:
1. Sensor AAD-1123 connects with client ID "publisher" -> Connected successfully
2. Sensor BBB-3423 connects with client ID "publisher" -> Kicks off AAD-1123
3. AAD-1123 reconnects with client ID "publisher" -> Kicks off BBB-3423
4. BBB-3423 reconnects with client ID "publisher" -> Kicks off AAD-1123
... (cycle continues)
```

This created a constant battle where sensors were disconnecting each other, resulting in:
- Only one sensor publishing at any given moment
- Lost measurements during disconnection/reconnection cycles
- RabbitMQ warnings: `MQTT disconnecting client with duplicate id 'publisher'`

## The Fix

Changed from a hardcoded client ID to using the unique sensor serial number:

### Before (Problematic Code)

```go
func NewConfig() (*Config, error) {
    // ...
    mqttOpts := MQTTCreateClientOptions("publisher", routing.RabbitMQTTConnString)
    // ...
}
```

Both sensors used the static client ID `"publisher"`.

### After (Fixed Code)

```go
func NewConfig(clientID string) (*Config, error) {
    // ...
    mqttOpts := MQTTCreateClientOptions(clientID, routing.RabbitMQTTConnString)
    // ...
}

func main() {
    serialNumber := os.Getenv("SENSOR_SERIAL_NUMBER")
    // ...
    cfg, err := NewConfig(serialNumber) // Use serial number as client ID
    // ...
}
```

Now each sensor uses its unique serial number:
- sensor-simulation-0: Client ID = `"AAD-1123"`
- sensor-simulation-1: Client ID = `"BBB-3423"`

## Best Practices for MQTT Client IDs

1. **Always use unique identifiers**: Device serial numbers, UUIDs, or MAC addresses
2. **Avoid hardcoded values**: Never use static strings like "publisher" or "client"
3. **Keep IDs consistent**: The same device should use the same client ID across reconnections
4. **Length limits**: Client IDs can be 1-23 characters (MQTT 3.1.1) or longer in MQTT 5.0
5. **Clean session flag**: Consider whether you need persistent sessions when a client reconnects

## How RabbitMQ Handles MQTT

RabbitMQ implements MQTT through a plugin that translates MQTT messages to AMQP:
- MQTT topics map to RabbitMQ exchanges and routing keys
- MQTT QoS levels map to AMQP delivery guarantees
- Client IDs are managed by the MQTT plugin's connection handler

When duplicate client IDs are detected, RabbitMQ logs warnings and enforces the MQTT specification by disconnecting the previous client.

## Verification

After the fix, you can verify both sensors are connected by checking RabbitMQ management UI:
1. Navigate to `http://localhost:15672` (default credentials: guest/guest)
2. Go to "Connections" tab
3. Look for connections with protocols "MQTT" showing both serial numbers as client IDs

Or check the logs:
```bash
docker logs iot-rabbitmq | grep "Accepted MQTT connection"
```

You should see two different client IDs connected without disconnection warnings.

## Related Issues

This issue also explains why:
- The sensor cache showed 2 sensors loaded but only 1 was receiving data
- Grafana queries for the second sensor showed minimal or no data
- RabbitMQ logs showed constant connect/disconnect cycles

## References

- [MQTT Version 3.1.1 Specification](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/mqtt-v3.1.1.html)
- [RabbitMQ MQTT Plugin Documentation](https://www.rabbitmq.com/mqtt.html)
- [Eclipse Paho MQTT Go Client](https://github.com/eclipse/paho.mqtt.golang)
