package main

import (
	"context"
	"log"
	"os"

	processor "github.com/mapofzones/txs-processor/pkg"
)

func main() {
	p, err := processor.NewProcessor(context.Background(), os.Getenv("rabbit"), "blocks")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(p.Process(context.Background()))
}
