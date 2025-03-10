package server

import (
	"context"

	"github.com/ekzyis/zapback/db"
	"github.com/ekzyis/zapback/lightning"
)

type Context struct {
	context.Context
	Env            string
	PublicURL      string
	CommitShortSha string
	CommitLongSha  string
	Ln             lightning.Lightning
	Db             *db.Db
}
