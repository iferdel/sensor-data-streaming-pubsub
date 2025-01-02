package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTableSensor() {

	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, routing.PostgresConnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	/********************************************/
	/* Create ordinary relational table         */
	/********************************************/

	queryCreateTable := `CREATE TABLE sensor (
		id SERIAL PRIMARY KEY, 
		serial_number VARCHAR(50),
	);`

	_, err = dbpool.Exec(ctx, queryCreateTable)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create `sensor` table: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully created relational table `sensor`")
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

	queryInsertMetadata := `INSERT INTO sensor (serial_number) VALUES ($1);`

	_, err = dbpool.Exec(ctx, queryInsertMetadata, serialNumber)
	if err != nil {
		return fmt.Errorf("Unable to insert sensor metadata into database: %v\n", err)
	}
	fmt.Printf("Inserted sensor (%s) into database \n", serialNumber)

	return nil
}
