package sensorlogic

import (
	"context"
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

// method from sensorstate maybe
func HandleMeasurements(ctx context.Context, db *storage.DB, dtos []routing.SensorMeasurement) error {

	sensorMap, err := db.GetSensorIDBySerialNumberMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch sensor IDs: %v", err)
	}

	records := make([]storage.SensorMeasurementRecord, len(dtos))
	for i, dto := range dtos {

		sensorID, exists := sensorMap[dto.SerialNumber]
		if !exists {
			return fmt.Errorf("sensor serial number not found: %s", dto.SerialNumber)
		}

		// Map DTO -to- DB Record
		records[i] = storage.SensorMeasurementRecord{
			Timestamp:   dto.Timestamp,
			SensorID:    sensorID,
			Measurement: dto.Value,
		}

	}

	if err := db.BatchArrayWriteMeasurement(ctx, records); err != nil {
		return fmt.Errorf("failed to write measurement: %v", err)
	}

	return nil
}
