package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	stridge "github.com/arshamhaq/stridge-goclient"
)

func main() {
	client, err := stridge.NewClient(stridge.Config{
		BaseURL: stridge.SandboxBaseURL,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	quote, err := client.Quote(ctx, stridge.QuoteRequest{
		FromNetworkID: 1,
		FromAsset:     "0x0000000000000000000000000000000000000000",
		ToNetworkID:   56,
		ToAsset:       "0x55d398326f99059fF775485246999027B3197955",
		Amount:        "1000000000000000000",
	})
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(quote); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Quote:", *quote)
}
