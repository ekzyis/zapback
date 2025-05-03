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
	GameCode       sql.NullString
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

type UpdateInvoice struct {
	ConfirmedAt sql.NullTime
}

func (tx *Tx) CreateInvoice(inv *CreateInvoice) (*Invoice, error) {
	// TODO: insert msats_received
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

func (db *Db) GetInvoice(paymentHash string) (*Invoice, error) {
	row := db.QueryRow(`
		SELECT
			invoice.id,
			invoice.created_at,
			invoice.expires_at,
			invoice.payment_hash,
			invoice.payment_request,
			invoice.confirmed_at,
			invoice.canceled_at,
			invoice.msats_requested,
			invoice.msats_received,
			invoice.description,
			invoice.player_id,
			invoice.game_id,
			game.code
		FROM invoice
		LEFT JOIN game ON invoice.game_id = game.id
		WHERE invoice.payment_hash = $1
	`, paymentHash)

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
		&invoice.GameCode,
	); err != nil {
		return nil, err
	}

	return &invoice, nil
}

func (tx *Tx) UpdateInvoice(id int, update *UpdateInvoice) error {
	_, err := tx.Exec(`
		UPDATE invoice
		SET confirmed_at = $2
		WHERE id = $1`,
		id, update.ConfirmedAt.Time,
	)
	return err
}

func (db *Db) GetPendingInvoices() ([]Invoice, error) {
	rows, err := db.Query(`
		SELECT
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
			game_id
		FROM invoice
		WHERE confirmed_at IS NULL AND expires_at > NOW()
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice
	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(
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
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (tx *Tx) GetGameInvoice(id int) (*Invoice, error) {
	row := tx.QueryRow(`
		SELECT
			invoice.id,
			invoice.created_at,
			invoice.expires_at,
			invoice.payment_hash,
			invoice.payment_request,
			invoice.confirmed_at,
			invoice.canceled_at,
			invoice.msats_requested,
			invoice.msats_received,
			invoice.description,
			invoice.player_id,
			invoice.game_id,
			game.code
		FROM invoice
		LEFT JOIN game ON invoice.game_id = game.id
		WHERE game_id = $1
		AND confirmed_at IS NULL
		ORDER BY invoice.created_at DESC
		LIMIT 1
	`, id)

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
		&invoice.GameCode,
	); err != nil {
		return nil, err
	}

	return &invoice, nil
}
