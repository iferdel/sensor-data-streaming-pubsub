package sensorlogic

import (
	"context"
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

// method from sensorstate maybe
func HandleMeasurements(ctx context.Context, db *storage.DB, dtos []routing.SensorMeasurement) error {
	for _, dto := range dtos {
		// inneficient query
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

		if err := db.WriteMeasurement(ctx, record); err != nil {
			return fmt.Errorf("failed to write measurement: %v", err)
		}
	}

	return nil
}
