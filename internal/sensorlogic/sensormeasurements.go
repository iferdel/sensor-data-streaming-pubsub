package sensorlogic

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

// method from sensorstate maybe
func HandleMeasurement(dto routing.SensorMeasurement) error {
	sensorID, err := storage.GetSensorIDBySerialNumber(dto.SerialNumber)
	if err != nil {
		return fmt.Errorf("failed to get sensor ID: %v", err)
	}

	// Map DTO -to- DB Record
	record := storage.SensorMeasurementRecord{
		Timestamp:   dto.Timestamp,
		SensorID:    sensorID,
		Measurement: dto.Value,
	}

	if err := storage.WriteMeasurement(record); err != nil {
		return fmt.Errorf("failed to write measurement: %v", err)
	}

	return nil
}
