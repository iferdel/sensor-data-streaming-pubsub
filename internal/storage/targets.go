package storage

import (
	"context"
	"fmt"
	"log"
)

func (DB *DB) GetTarget(ctx context.Context) ([]TargetRecord, error) {

	queryGetMetadata := `SELECT name FROM target;`

	rows, err := DB.pool.Query(ctx, queryGetMetadata)
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

func (DB *DB) WriteTarget(ctx context.Context, tr TargetRecord) error {
	// TODO: Implement Mutex RW

	// if target exists, return log message with kind of 'target already registered'
	queryCheckIfExists := `SELECT EXISTS (
		SELECT 1 FROM target WHERE name = ($1)
	);`

	var rowExists bool
	err := DB.pool.QueryRow(ctx, queryCheckIfExists, tr.Name).Scan(&rowExists)
	if err != nil {
		log.Fatal(err)
	}

	if rowExists {
		fmt.Printf("Entry for target `%s` already exists. Skipping...\n", tr.Name)
		return nil
	}

	queryInsertMetadata := `INSERT INTO target (name) VALUES ($1);`

	_, err = DB.pool.Exec(ctx, queryInsertMetadata, tr.Name)
	if err != nil {
		return fmt.Errorf("unable to insert target metadata into database: %v", err)
	}
	fmt.Printf("Inserted target (%s) into `target` table\n", tr.Name)

	return nil
}
