package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTableSensor() error {

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* Create ordinary relational table         */
	/********************************************/

	queryCheckIfExists := `SELECT EXISTS (
		SELECT FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename = 'sensor'
	);`

	var tableExists bool
	err = dbpool.QueryRow(ctx, queryCheckIfExists).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("Error checking existency of `sensor` table: %v", err)
	}

	if tableExists {
		fmt.Println("Table `sensor` already exists. Skipping...")
		return nil
	}

	queryCreateTable := `CREATE TABLE sensor (
		id SERIAL PRIMARY KEY, 
		serial_number VARCHAR(50) UNIQUE NOT NULL
	);`

	_, err = dbpool.Exec(ctx, queryCreateTable)

	if err != nil {
		return fmt.Errorf("Unable to create `sensor` table: %v\n", err)
	}
	fmt.Println("Successfully created relational table `sensor`")

	return nil
}

func GetSensor() error {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* SELECT from relational table             */
	/********************************************/

	queryGetMetadata := `SELECT serial_number FROM sensor;`

	rows, err := dbpool.Query(ctx, queryGetMetadata)
	if err != nil {
		return fmt.Errorf("Unable to get sensors: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var serialNumber string
		err = rows.Scan(&serialNumber)
		if err != nil {
			return err
		}
		fmt.Println(serialNumber)
	}
	return nil
}

func WriteSensor(serialNumber string) error {

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
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
	err = dbpool.QueryRow(ctx, queryCheckIfExists, serialNumber).Scan(&rowExists)
	if err != nil {
		log.Fatal(err)
	}

	if rowExists {
		fmt.Printf("Entry for sensor `%s` already exists. Skipping...\n", serialNumber)
		return nil
	}

	queryInsertMetadata := `INSERT INTO sensor (serial_number) VALUES ($1);`

	_, err = dbpool.Exec(ctx, queryInsertMetadata, serialNumber)
	if err != nil {
		return fmt.Errorf("Unable to insert sensor metadata into database: %v\n", err)
	}
	fmt.Printf("Inserted sensor (%s) into `sensor` table\n", serialNumber)

	return nil
}

func DeleteSensor(serialNumber string) error {

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
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
		return fmt.Errorf("Unable to delete sensor metadata from database: %v\n", err)
	}
	fmt.Printf("Deleted sensor (%s) from `sensor` table (and all its measurements)\n", serialNumber)
	return nil
}
