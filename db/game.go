package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Game struct {
	Id        int
	CreatedAt time.Time
	ZapAmount int64
	Code      string
}

type CreateGame struct {
	ZapAmount int64
}

type Turn struct {
	// player who is taking their turn
	Player Player
	// total pool amount of game
	PoolAmount int64
	// the invoice that the player must pay to take their turn
	Invoice *Invoice
}

type GameStatus struct {
	Id         int
	CreatedAt  time.Time
	Code       string
	PoolAmount int64
	Status     string
	Expired    bool
}

func (tx *Tx) CreateGame(g *CreateGame) (*Game, error) {
	row := tx.QueryRow(`
		INSERT INTO game (zap_amount)
		VALUES ($1)
		RETURNING id, created_at, zap_amount, code`,
		g.ZapAmount,
	)

	var game Game
	if err := row.Scan(&game.Id, &game.CreatedAt, &game.ZapAmount, &game.Code); err != nil {
		return &Game{}, err
	}

	return &game, nil
}

func (db *Db) GetGame(code string) (*Game, error) {
	row := db.QueryRow(`
		SELECT id, created_at, zap_amount, code
		FROM game
		WHERE code = $1`,
		code,
	)

	var game Game
	if err := row.Scan(&game.Id, &game.CreatedAt, &game.ZapAmount, &game.Code); err != nil {
		return nil, err
	}

	return &game, nil
}

func (tx *Tx) HasGameStarted(id int) (bool, error) {
	row := tx.QueryRow(`
		SELECT COUNT(*)
		FROM invoice
		WHERE game_id = $1
		AND confirmed_at IS NOT NULL
	`, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count >= 2, nil
}

func (tx *Tx) GetGameTurn(id int) (*Turn, error) {
	row := tx.QueryRow(`
		SELECT player.id, player.lnaddr
		FROM invoice
		LEFT JOIN player ON invoice.player_id = player.id
		WHERE invoice.game_id = $1
		AND invoice.confirmed_at IS NOT NULL
		ORDER BY invoice.created_at DESC
		OFFSET 1
		LIMIT 1
	`, id)

	var player Player
	if err := row.Scan(&player.Id, &player.LightningAddress); err != nil {
		return nil, err
	}

	row = tx.QueryRow(`
		SELECT SUM(invoice.msats_requested)
		FROM invoice
		WHERE game_id = $1
		AND confirmed_at IS NOT NULL
	`, id)

	var poolAmount int64
	if err := row.Scan(&poolAmount); err != nil {
		return nil, err
	}

	inv, err := tx.GetGameInvoice(id)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get game invoice: %s", err.Error())
		}
	}

	return &Turn{Player: player, PoolAmount: poolAmount, Invoice: inv}, nil
}

func (tx *Tx) GetGameWinner(id int) (*Player, error) {
	row := tx.QueryRow(`
		SELECT player.id, player.lnaddr
		FROM invoice
		LEFT JOIN player ON invoice.player_id = player.id
		WHERE invoice.game_id = $1
		AND invoice.confirmed_at IS NOT NULL
		ORDER BY invoice.created_at DESC
		LIMIT 1
	`, id)

	var player Player
	if err := row.Scan(&player.Id, &player.LightningAddress); err != nil {
		return nil, err
	}

	return &player, nil
}

func (db *Db) GetGameStats() ([]*GameStatus, error) {
	rows, err := db.Query(`
		WITH
			pool AS (
				SELECT
					game.id,
					COALESCE(
						SUM(invoice.msats_requested) FILTER (WHERE invoice.confirmed_at IS NOT NULL),
						0
					) as pool_amount
				FROM game
				LEFT JOIN invoice ON invoice.game_id = game.id
				GROUP BY game.id
			),
			latest_invoice AS (
				SELECT DISTINCT ON (i.game_id)
					i.*
				FROM invoice i
				ORDER BY i.game_id, i.created_at DESC
			),
			status AS (
				SELECT
					g.id, g.created_at, g.code,
					p.pool_amount,
					li.type AS status,
					li.expires_at <= NOW() as expired
				FROM game g
				LEFT JOIN pool p ON p.id = g.id
				LEFT JOIN latest_invoice li ON li.game_id = g.id
				ORDER BY g.id DESC
			)
		SELECT id, created_at, code, pool_amount, status, expired
		FROM status
	`)
	if err != nil {
		return nil, err
	}

	var stats []*GameStatus
	for rows.Next() {
		var status GameStatus
		if err := rows.Scan(&status.Id, &status.CreatedAt, &status.Code, &status.PoolAmount, &status.Status, &status.Expired); err != nil {
			return nil, err
		}
		stats = append(stats, &status)
	}

	return stats, nil
}
