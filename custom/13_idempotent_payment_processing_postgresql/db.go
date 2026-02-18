package main

import (
	"database/sql"
	"log"
)

type DB struct {
	conn *sql.DB
}

func NewDB() *DB {
	db, err := sql.Open("postgres",
		"host=localhost user=postgres password=postgres dbname=payments_test port=5444 sslmode=disable")
	if err != nil {
		log.Fatalf("err opening db: %v", err)
	}

	if err := createTables(db); err != nil {
		log.Fatalf("err opening db: %v", err)
	}

	return &DB{
		conn: db,
	}
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
		CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed');
	END IF;
END $$;

CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status payment_status NOT NULL,
    stripe_charge_id VARCHAR(255),
    error_message VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_idempotency_key ON payments(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_user_id ON payments(user_id);
	`)
	if err != nil {
		return err
	}
	return nil
}
