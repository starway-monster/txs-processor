package main

import (
	"context"
	"log"
	"os"

	processor "github.com/mapofzones/txs-processor/pkg"
)

func main() {
	p, err := processor.NewProcessor(context.Background(), os.Getenv("rabbit"), "block")
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = p.Process(ctx)
	cancel()
	log.Fatal(err)
}
