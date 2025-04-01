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
	query += ";"

	_, err := DB.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to insert batch of sensor measurements into Timescale %v", err)
	}
	fmt.Printf("%v - Successfully inserted batches of %v into `measurement` hypertable", time.Now(), len(measurements))

	return nil
}
