package cmd

//import "fmt"

//func main() {
//   fmt.Println("sending awake command to sensor", "sensorSerialNumber")
//err := pubsub.PublishGob(
//    publishCh,                // amqp.Channel
//    routing.ExchangeTopicIoT, // exchange
//    fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
//    routing.CommandMessage{
//        SensorName: sensorSerialNumber,
//        Timestamp:  time.Now(),
//        Command:    "awake",
//        Params:     nil,
//    }, // value
//)
//if err != nil {
//    log.Printf("could not publish awake command: %v", err)
//}
//}
