package main

import (
	"log"
	"os"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/dperez9/cert-manager-webhook-duckdns/duckdns"
)

func main() {
	GroupName := os.Getenv("GROUP_NAME")
	if GroupName == "" {
		log.Fatal("GROUP_NAME must be specified")
	}

	// Registra el solver DuckDNS en el webhook, con el nombre de grupo proporcionado.
	cmd.RunWebhookServer(GroupName, duckdns.NewSolver())
}
