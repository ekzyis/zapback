package db

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Db struct {
	*sql.DB
}

type Tx struct {
	*sql.Tx
}

func New(url string) (*Db, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &Db{db}, nil
}

func (d *Db) Migrate() error {
	if _, err := d.Exec(`
		CREATE EXTENSION IF NOT EXISTS pgcrypto
	`); err != nil {
		return err
	}

	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS game (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			zap_amount BIGINT NOT NULL,
			code TEXT NOT NULL DEFAULT encode(gen_random_bytes(8), 'hex') UNIQUE
		)
	`); err != nil {
		return err
	}

	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS player (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			lnaddr TEXT NOT NULL,
			game_id INTEGER NOT NULL REFERENCES game(id)
		)
	`); err != nil {
		return err
	}

	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS invoice (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMPTZ NOT NULL,
			payment_hash TEXT NOT NULL,
			payment_request TEXT NOT NULL,
			confirmed_at TIMESTAMPTZ,
			canceled_at TIMESTAMPTZ,
			msats_requested BIGINT NOT NULL,
			msats_received BIGINT,
			description TEXT,
			player_id INTEGER NOT NULL REFERENCES player(id),
			game_id INTEGER NOT NULL REFERENCES game(id)
		)
	`); err != nil {
		return err
	}

	return nil
}

func (d *Db) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}
