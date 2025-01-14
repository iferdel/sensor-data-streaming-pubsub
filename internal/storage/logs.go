package storage

import (
	"fmt"
	"os"
	"time"
)

const logPath = "log/iot.log"

func WriteLog(sensorLog SensorLogRecord) error {
	fmt.Printf("received logs from %v...\n", sensorLog.SerialNumber)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open logs file: %v", err)
	}
	defer f.Close()
	formattedTime := sensorLog.Timestamp.Format(time.RFC3339Nano)

	str := fmt.Sprintf("time=\"%s\" logger=\"%v\" level=\"%v\" message=\"%v\"\n",
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
