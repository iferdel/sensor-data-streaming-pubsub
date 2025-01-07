/*
- Uses TimescaleDB for storing measurements in hypertables.
- TimescaleDB tables do not support primary keys.
- Contains CRUD operations for sensor_measurement table.
*/
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func WriteMeasurement(measurement SensorMeasurementRecord) error {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* INSERT into hypertable                   */
	/********************************************/

	queryInsertTimeseriesData := `INSERT INTO sensor_measurement (time, sensor_id, measurement) VALUES ($1, $2, $3);`

	_, err = dbpool.Exec(ctx, queryInsertTimeseriesData, measurement.Timestamp, measurement.SensorID, measurement.Measurement)
	if err != nil {
		return fmt.Errorf("Unable to insert sample into Timescale %v\n", err)
	}
	fmt.Printf("%v - Successfully inserted sample into `measurement` hypertable", time.Now())
	// TODO: as many inserts as rows of data, the idea is to deploy it with this pattern, measure the way the whole system behaves (broker, backend, db) and then optmize with batch processing

	return nil
}
