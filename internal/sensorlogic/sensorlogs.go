package sensorlogic

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

// method from sensorstate maybe
func HandleLogs(dto routing.SensorLog) error {
	// Map DTO -to- DB Record
	record := storage.SensorLogRecord{
		Timestamp:    dto.Timestamp,
		SerialNumber: dto.SerialNumber,
		Level:        dto.Level,
		Message:      dto.Message,
	}

	if err := storage.WriteLog(record); err != nil {
		return fmt.Errorf("failed to write log: %v", err)
	}
	return nil
}
