package main

import (
	"fmt"
	"log"

	"github.com/ekzyis/zapback/env"
	"github.com/ekzyis/zapback/lightning"
	"github.com/ekzyis/zapback/server"
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatalf("error loading env: %v", err)
	}
	env.Parse()

	log.Printf("url:      %s", env.PublicUrl)
	log.Printf("commit:   %s", env.CommitShortSha)
	log.Printf("phoenixd: %s", env.PhoenixdUrl)

	p := lightning.NewPhoenixd(
		lightning.WithPhoenixdUrl(env.PhoenixdUrl),
		lightning.WithPhoenixdLimitedAccessToken(env.PhoenixdLimitedAccessToken),
	)

	s := server.New(server.Context{
		Env:            env.Env,
		PublicURL:      env.PublicUrl,
		CommitShortSha: env.CommitShortSha,
		CommitLongSha:  env.CommitLongSha,
		Ln:             p,
	})

	if err := s.Start(fmt.Sprintf(":%d", env.Port)); err != nil {
		log.Fatal(err)
	}
}
