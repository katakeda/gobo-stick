package main

import (
	"database/sql"
	"fmt"
	"os"
)

// DB ...
type DB struct {
	client *sql.DB
}

func (db *DB) init() error {
	client, err := sql.Open(os.Getenv("DB_DRIVER"), fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	))
	if err != nil {
		return err
	}
	db.client = client

	return nil
}

var db DB
