package db

import (
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
