package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTableMeasurement() {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
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
		log.Fatal(err)
	}

	if tableExists {
		fmt.Println("Table `sensor_measurement` already exists.")
		return
	}

	queryCreateTable := `CREATE TABLE sensor_measurement (
		time TIMESTAMPZ NOT NULL,
		sensor_id INTEGER,
		measurement DOUBLE PRECISION,
		UNIQUE (time, sensor_id),
		CONSTRAINT fk_sensor 
			FOREIGN KEY (sensor_id) 
				REFERENCES sensor(id) 
					ON DELETE CASCADE
	);`

	queryCreateHyperTable := `SELECT create_hypertable('sensor_measurement', by_range(time));`

	_, err = dbpool.Exec(ctx, queryCreateTable+queryCreateHyperTable)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create `sensor_measurement` hypertable: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully created hypertable `sensor_measurement`")
}

type measurement struct {
	Time        time.Time
	SensorId    int
	Measurement float64
}

func WriteMeasurement(measurements []measurement) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	/********************************************/
	/* INSERT into hypertable                   */
	/********************************************/

	queryInsertTimeseriesData := `
		INSERT INTO measurements (time, sensor_id, measurement) values ($1, $2, $3);
	`

	for i := range measurements {
		var m measurement
		m = measurements[i]
		_, err := dbpool.Exec(ctx, queryInsertTimeseriesData, m.Time, m.SensorId, m.Measurement)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to insert sample into Timescale %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully inserted samples into `measurement` hypertable")
	}
	// TODO: as many inserts as rows of data, the idea is to deploy it with this pattern, measure the way the whole system behaves (broker, backend, db) and then optmize with batch processing
}
