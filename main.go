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
	log.Printf("voltage:  %s", env.VoltageUrl)

	v := lightning.NewVoltage(
		lightning.WithVoltageUrl(env.VoltageUrl),
		lightning.WithVoltageOrganizationId(env.VoltageOrganizationId),
		lightning.WithVoltageEnvId(env.VoltageEnvId),
		lightning.WithVoltageWalletId(env.VoltageWalletId),
		lightning.WithVoltageApiKey(env.VoltageApiKey),
	)

	s := server.New(server.Context{
		Env:            env.Env,
		PublicURL:      env.PublicUrl,
		CommitShortSha: env.CommitShortSha,
		CommitLongSha:  env.CommitLongSha,
		Ln:             v,
	})

	if err := s.Start(fmt.Sprintf(":%d", env.Port)); err != nil {
		log.Fatal(err)
	}
}
