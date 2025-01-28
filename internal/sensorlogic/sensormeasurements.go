package sensorlogic

import (
	"context"
	"fmt"

	"github.com/iferdel/treanteyes/internal/routing"
	"github.com/iferdel/treanteyes/internal/storage"
)

// method from sensorstate maybe
func HandleMeasurement(ctx context.Context, db *storage.DB, dto routing.SensorMeasurement) error {
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

	return nil
}
