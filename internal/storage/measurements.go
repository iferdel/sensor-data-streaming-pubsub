/*
- Uses TimescaleDB for storing measurements in hypertables.
- TimescaleDB tables do not support primary keys.
- Contains CRUD operations for sensor_measurement table.
*/
package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (DB *DB) WriteMeasurement(ctx context.Context, measurement SensorMeasurementRecord) error {
	/********************************************/
	/* INSERT into hypertable                   */
	/********************************************/

	queryInsertTimeseriesData := `INSERT INTO sensor_measurement (time, sensor_id, measurement) VALUES ($1, $2, $3);`

	_, err := DB.pool.Exec(ctx, queryInsertTimeseriesData, measurement.Timestamp, measurement.SensorID, measurement.Measurement)
	if err != nil {
		return fmt.Errorf("unable to insert sample into Timescale %v", err)
	}
	fmt.Printf("%v - Successfully inserted sample into `measurement` hypertable", time.Now())
	// TODO: as many inserts as rows of data, the idea is to deploy it with this pattern, measure the way the whole system behaves (broker, backend, db) and then optmize with batch processing

	return nil
}

func (DB *DB) BatchWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error {
	/********************************************/
	/* INSERT into hypertable                   */
	/********************************************/

	query := `INSERT INTO sensor_measurement (time, sensor_id, measurement) VALUES `
	var args []interface{}
	valueStrings := []string{}

	for i, measurement := range measurements {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, measurement.Timestamp, measurement.SensorID, measurement.Measurement)
	}

	query += strings.Join(valueStrings, ", ")
	query += ` ON CONFLICT (time, sensor_id) DO NOTHING;`

	_, err := DB.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to insert batch of sensor measurements into Timescale %v", err)
	}
	fmt.Printf("%v - Successfully inserted batches of %v into `measurement` hypertable", time.Now(), len(measurements))

	return nil
}

func (DB *DB) BatchArrayWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error {

	if len(measurements) == 0 {
		return nil
	}

	timeSlice := make([]time.Time, len(measurements))
	sensorIDSlice := make([]int, len(measurements))
	measurementSlice := make([]float64, len(measurements))

	for i, measurement := range measurements {
		timeSlice[i] = measurement.Timestamp
		sensorIDSlice[i] = measurement.SensorID
		measurementSlice[i] = measurement.Measurement
	}

	query := `
		INSERT INTO sensor_measurement (time, sensor_id, measurement)
			SELECT * 
			FROM unnest(
				$1::timestamptz[],
				$2::int[],
				$3::double precision[]
			) AS t(time, sensor_id, measurement)
			ON CONFLICT (sensor_id, time) DO NOTHING;
		`

	batchSize := 25_000
	total := len(measurements)

	for start := 0; start < total; start += batchSize {
		end := start + batchSize
		if end > total {
			end = total
		}

		timeBatch := timeSlice[start:end]
		sensorIDBatch := sensorIDSlice[start:end]
		measurementBatch := measurementSlice[start:end]

		_, err := DB.pool.Exec(ctx, query, timeBatch, sensorIDBatch, measurementBatch)
		if err != nil {
			return fmt.Errorf("failed inserting batch [%d:%d]: %w", start, end, err)
		}
	}

	fmt.Printf("%v - Successfully inserted batches of %v into `measurement` hypertable", time.Now(), len(measurements))

	return nil
}

func (DB *DB) CopyWriteMeasurement(ctx context.Context, measurements []SensorMeasurementRecord) error {
	copyCount, err := DB.pool.CopyFrom(
		ctx,
		pgx.Identifier{"sensor_measurement"},
		[]string{"time", "sensor_id", "measurement"},
		pgx.CopyFromSlice(
			len(measurements), func(i int) ([]interface{}, error) {
				return []interface{}{
					measurements[i].Timestamp,
					measurements[i].SensorID,
					measurements[i].Measurement,
				}, nil
			}),
	)
	if err != nil {
		return fmt.Errorf("unable to COPY sensor measurements into Timescale %v", err)
	}
	fmt.Printf("%v - Successfully COPY (%d) points into `measurement` hypertable", time.Now(), copyCount)

	return nil
}
