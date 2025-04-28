package db

import (
	"database/sql"
	"time"
)

type Invoice struct {
	Id             int
	CreatedAt      time.Time
	ExpiresAt      time.Time
	PaymentHash    string
	PaymentRequest string
	ConfirmedAt    sql.NullTime
	CanceledAt     sql.NullTime
	MsatsRequested int64
	MsatsReceived  sql.NullInt64
	Description    sql.NullString
	PlayerId       int
	GameId         int
}

type CreateInvoice struct {
	CreatedAt      time.Time
	ExpiresAt      time.Time
	PaymentHash    string
	PaymentRequest string
	MsatsRequested int64
	Description    string
	PlayerId       int
	GameId         int
}

func (tx *Tx) CreateInvoice(inv *CreateInvoice) (*Invoice, error) {
	row := tx.QueryRow(`
			INSERT INTO invoice (
				created_at,
				expires_at,
				payment_hash,
				payment_request,
				msats_requested,
				description,
				player_id,
				game_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING
				id,
				created_at,
				expires_at,
				payment_hash,
				payment_request,
				confirmed_at,
				canceled_at,
				msats_requested,
				msats_received,
				description,
				player_id,
				game_id`,
		inv.CreatedAt,
		inv.ExpiresAt,
		inv.PaymentHash,
		inv.PaymentRequest,
		inv.MsatsRequested,
		inv.Description,
		inv.PlayerId,
		inv.GameId,
	)

	var invoice Invoice
	if err := row.Scan(
		&invoice.Id,
		&invoice.CreatedAt,
		&invoice.ExpiresAt,
		&invoice.PaymentHash,
		&invoice.PaymentRequest,
		&invoice.ConfirmedAt,
		&invoice.CanceledAt,
		&invoice.MsatsRequested,
		&invoice.MsatsReceived,
		&invoice.Description,
		&invoice.PlayerId,
		&invoice.GameId,
	); err != nil {
		return &Invoice{}, err
	}

	return &invoice, nil
}
