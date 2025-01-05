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

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTableMeasurement() error {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* Create hypertable table                  */
	/********************************************/

	queryCheckIfExists := `SELECT EXISTS (
		SELECT FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename = 'sensor_measurement'
	);`

	var tableExists bool
	err = dbpool.QueryRow(ctx, queryCheckIfExists).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("Error checking existency of `sensor_measurement` table: %v", err)
	}

	if tableExists {
		fmt.Println("Table `sensor_measurement` already exists.")
		return nil
	}

	queryCreateTable := `CREATE TABLE sensor_measurement (
		time TIMESTAMPTZ NOT NULL,
		sensor_id INTEGER,
		measurement DOUBLE PRECISION,
		UNIQUE (time, sensor_id),
		CONSTRAINT fk_sensor 
			FOREIGN KEY (sensor_id) 
				REFERENCES sensor(id) 
					ON DELETE CASCADE
	);`

	queryCreateHyperTable := `SELECT create_hypertable('sensor_measurement', by_range('time'));`

	_, err = dbpool.Exec(ctx, queryCreateTable+queryCreateHyperTable)

	if err != nil {
		return fmt.Errorf("Unable to create `sensor_measurement` hypertable: %v\n", err)

	}
	fmt.Println("Successfully created hypertable `sensor_measurement`")
	return nil
}

func WriteMeasurement(measurement routing.SensorMeasurement) error {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* INSERT into hypertable                   */
	/********************************************/

	queryInsertTimeseriesData := `
		INSERT INTO sensor_measurement (time, sensor_id, measurement) 
			VALUES (
				$1, 
				(SELECT id FROM sensor WHERE serial_number = $2),
				$3
		);`

	_, err = dbpool.Exec(ctx, queryInsertTimeseriesData, measurement.Timestamp, measurement.SerialNumber, measurement.Value)
	if err != nil {
		return fmt.Errorf("Unable to insert sample into Timescale %v\n", err)
	}
	fmt.Printf("%v - Successfully inserted sample into `measurement` hypertable", time.Now())
	// TODO: as many inserts as rows of data, the idea is to deploy it with this pattern, measure the way the whole system behaves (broker, backend, db) and then optmize with batch processing

	return nil
}
