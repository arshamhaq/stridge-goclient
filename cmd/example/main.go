package main

import (
	"context"
	"log"
	"os"
	"time"

	stridge "github.com/arshamhaq/stridge-goclient"
)

func main() {
	client, err := stridge.NewClient(stridge.Config{
		BaseURL:    stridge.SandboxBaseURL,
		APIKey:     os.Getenv("STRIDGE_API_KEY"),
		GatewayKey: os.Getenv("STRIDGE_GATEWAY_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TODO: Build a QuoteRequest and call client.Quote(ctx, request) after the
	// shared request infrastructure and Quote endpoint have been implemented.
	_ = client
	_ = ctx
}
