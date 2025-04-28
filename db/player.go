package db

import "time"

type Player struct {
	Id               int
	CreatedAt        time.Time
	LightningAddress string
	GameId           int
}

type CreatePlayer struct {
	LightningAddress string
	GameId           int
}

func (tx *Tx) CreatePlayer(p *CreatePlayer) (*Player, error) {
	row := tx.QueryRow(`
		INSERT INTO player (lnaddr, game_id)
		VALUES ($1, $2)
		RETURNING id, created_at, lnaddr, game_id`,
		p.LightningAddress,
		p.GameId,
	)

	var player Player
	if err := row.Scan(
		&player.Id,
		&player.CreatedAt,
		&player.LightningAddress,
		&player.GameId,
	); err != nil {
		return &Player{}, err
	}

	return &player, nil
}
