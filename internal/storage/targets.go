package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetTarget() ([]TargetRecord, error) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, PostgresConnString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	/********************************************/
	/* SELECT from relational table             */
	/********************************************/

	queryGetMetadata := `SELECT name FROM target;`

	rows, err := dbpool.Query(ctx, queryGetMetadata)
	if err != nil {
		return nil, fmt.Errorf("unable to get targets: %v", err)
	}
	defer rows.Close()

	var targets []TargetRecord

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		targets = append(targets, TargetRecord{
			Name: name,
		})
	}

	return targets, nil
}

func WriteTarget(tr TargetRecord) error {
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

	// if target exists, return log message with kind of 'target already registered'
	queryCheckIfExists := `SELECT EXISTS (
		SELECT 1 FROM target WHERE name = ($1)
	);`

	var rowExists bool
	err = dbpool.QueryRow(ctx, queryCheckIfExists, tr.Name).Scan(&rowExists)
	if err != nil {
		log.Fatal(err)
	}

	if rowExists {
		fmt.Printf("Entry for target `%s` already exists. Skipping...\n", tr.Name)
		return nil
	}

	queryInsertMetadata := `INSERT INTO target (name) VALUES ($1);`

	_, err = dbpool.Exec(ctx, queryInsertMetadata, tr.Name)
	if err != nil {
		return fmt.Errorf("unable to insert target metadata into database: %v", err)
	}
	fmt.Printf("Inserted target (%s) into `target` table\n", tr.Name)

	return nil
}
