package storage

import (
	"fmt"
	"os"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
)

const logsFile = "iot.log"

func WriteLog(sensorLog routing.SensorLog) error {
	fmt.Printf("received logs from %v...\n", sensorLog.SerialNumber)

	f, err := os.OpenFile(logsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open logs file: %v", err)
	}
	defer f.Close()
	formattedTime := sensorLog.TimeStamp.Format(time.RFC3339Nano)

	str := fmt.Sprintf("%s %v (%v): %v\n",
		formattedTime,
		sensorLog.SerialNumber,
		sensorLog.Level,
		sensorLog.Message,
	)

	_, err = f.WriteString(str)
	if err != nil {
		return fmt.Errorf("could not write to logs file from sensor %v: %v", sensorLog.SerialNumber, err)
	}

	return nil
}
