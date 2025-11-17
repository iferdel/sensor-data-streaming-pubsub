package sensorlogic

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

type SensorCache struct {
	mu          sync.RWMutex
	mapping     map[string]int
	lastRefresh time.Time
	db          *storage.DB
}

func NewSensorCache(ctx context.Context, db *storage.DB) (*SensorCache, error) {
	cache := &SensorCache{
		db:      db,
		mapping: make(map[string]int),
	}

	// initial fetch (refresh is stated for refreshing, but it could be used here to state it on init)
	err := cache.refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed initial sensor cache load: %v", err)
	}

	return cache, nil
}

func (sc *SensorCache) refresh(ctx context.Context) error {
	sensorMap, err := sc.db.GetSensorIDBySerialNumberMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch sensor IDs: %v", err)
	}
	sc.mu.Lock()
	sc.mapping = sensorMap
	sc.lastRefresh = time.Now()
	sc.mu.Unlock()

	fmt.Printf("[%s] Sensor cache refreshed: %d sensors loaded\n", time.Now().Format(time.RFC3339), len(sensorMap))

	return nil
}

func (sc *SensorCache) StartRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := sc.refresh(ctx)
			if err != nil {
				fmt.Printf("ERROR: sensor cache refresh failed: %v\n", err)
			}
		case <-ctx.Done():
			fmt.Println("Sensor cache refresh loop stopped")
			return
		}
	}
}

func (sc *SensorCache) Get(serialNumber string) (int, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	id, exists := sc.mapping[serialNumber]
	return id, exists
}

func (sc *SensorCache) GetAll() map[string]int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	mapCopy := make(map[string]int, len(sc.mapping)) // to avoid race conditions
	maps.Copy(mapCopy, sc.mapping)
	return mapCopy
}

func HandleMeasurementsWithCache(ctx context.Context, cache *SensorCache, db *storage.DB, dtos []routing.SensorMeasurement) error {
	sensorMap := cache.GetAll()

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
