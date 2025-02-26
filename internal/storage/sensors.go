/*
- Contains CRUD operations for sensor table.
*/
package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetSensorIDBySerialNumber(serialNumber string) (sensorID int, err error) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return 0, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	queryGetSensor := `
		SELECT id 
		FROM sensor
		WHERE serial_number = ($1)
	;`

	err = dbpool.QueryRow(ctx, queryGetSensor, serialNumber).Scan(&sensorID)
	if err != nil {
		return 0, fmt.Errorf("unable to query sensor ID: %v", err)
	}

	return sensorID, nil
}

func GetSensorBySerialNumber(serialNumber string) (sensor SensorRecord, err error) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return SensorRecord{}, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	queryGetSensor := `
		SELECT serial_number, sample_frequency 
		FROM sensor
		WHERE serial_number = ($1)
	;`

	err = dbpool.QueryRow(ctx, queryGetSensor, serialNumber).Scan(
		&sensor.SerialNumber,
		&sensor.SampleFrequency,
	)
	if err != nil {
		return SensorRecord{}, fmt.Errorf("unable to query sensor: %v", err)
	}

	return sensor, nil
}

func GetSensor() ([]SensorRecord, error) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* SELECT from relational table             */
	/********************************************/

	queryGetMetadata := `SELECT serial_number FROM sensor;`

	rows, err := dbpool.Query(ctx, queryGetMetadata)
	if err != nil {
		return nil, fmt.Errorf("unable to get sensors: %v", err)
	}
	defer rows.Close()

	var sensors []SensorRecord

	for rows.Next() {
		var serialNumber string
		err = rows.Scan(&serialNumber)
		if err != nil {
			return nil, err
		}

		sensors = append(sensors, SensorRecord{
			SerialNumber: serialNumber,
		})
	}

	return sensors, nil
}

func WriteSensor(sr SensorRecord) error {
	// TODO: Implement Mutex RW

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* INSERT into relational table             */
	/********************************************/

	// if sensor exists, return log message with kind of 'sensor already registered'
	queryCheckIfExists := `SELECT EXISTS (
		SELECT 1 FROM sensor WHERE serial_number = ($1)
	);`

	var rowExists bool
	err = dbpool.QueryRow(ctx, queryCheckIfExists, sr.SerialNumber).Scan(&rowExists)
	if err != nil {
		log.Fatal(err)
	}

	if rowExists {
		fmt.Printf("Entry for sensor `%s` already exists. Skipping...\n", sr.SerialNumber)
		return nil
	}

	queryInsertMetadata := `INSERT INTO sensor (serial_number, sample_frequency) VALUES ($1, $2);`

	_, err = dbpool.Exec(ctx, queryInsertMetadata, sr.SerialNumber, sr.SampleFrequency)
	if err != nil {
		return fmt.Errorf("unable to insert sensor metadata into database: %v", err)
	}
	fmt.Printf("Inserted sensor (%s) into `sensor` table\n", sr.SerialNumber)

	return nil
}

func DeleteSensor(serialNumber string) error {

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* DELETE relational table             			*/
	/********************************************/

	// if sensor does not exist, return log message with kind of 'sensor not registered'
	queryCheckIfExists := `SELECT EXISTS (
		SELECT 1 FROM sensor WHERE serial_number = ($1)
	);`

	var rowExists bool
	err = dbpool.QueryRow(ctx, queryCheckIfExists, serialNumber).Scan(&rowExists)
	if err != nil {
		log.Fatal(err)
	}

	if !rowExists {
		fmt.Printf("Entry for sensor `%s` does not exist. Skipping...\n", serialNumber)
		return nil
	}

	queryDeleteMetadata := `DELETE FROM sensor WHERE serial_number = ($1);`

	_, err = dbpool.Exec(ctx, queryDeleteMetadata, serialNumber)
	if err != nil {
		return fmt.Errorf("unable to delete sensor metadata from database: %v", err)
	}
	fmt.Printf("Deleted sensor (%s) from `sensor` table (and all its measurements)\n", serialNumber)
	return nil
}
