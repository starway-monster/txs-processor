package main

import (
	"context"
	"flag"
	"log"

	processor "github.com/mapofzones/txs-processor/pkg"
)

var graphql *string
var rabbit *string

func init() {
	graphql = flag.String("graphql", "localhost/v1/example", "endpoint for graphql")
	rabbit = flag.String("rabbit", "localhost/example", "rabbitMQ endpoint")
}

func main() {
	flag.Parse()

	p, err := processor.NewProcessor(context.Background(), *rabbit, "blocks", *graphql)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(p.Process(context.Background()))
}
