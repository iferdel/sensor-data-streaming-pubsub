package cmd

//
//			if commandInput[1] == "" {
//				fmt.Println("sample frequency must have an argument for the value")
//				continue
//			}
//			value, err := strconv.Atoi(commandInput[1])
//			if err != nil {
//				fmt.Println("sample frequency must be an integer greater than 0")
//				continue
//			}
//			fmt.Println("sending change sample frequency command to sensor", sensorSerialNumber)
//			err = pubsub.PublishGob(
//				publishCh,                // amqp.Channel
//				routing.ExchangeTopicIoT, // exchange
//				fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
//				routing.CommandMessage{
//					SensorName: sensorSerialNumber,
//					Timestamp:  time.Now(),
//					Command:    "changeSampleFrequency",
//					Params: map[string]interface{}{
//						"sampleFrequency": value,
//					},
//				}, // value
//			)
//			if err != nil {
//				log.Printf("could not publish change sample frequency command: %v", err)
//			}
