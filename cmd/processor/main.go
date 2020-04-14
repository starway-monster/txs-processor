package main

import (
	"context"
	"flag"
	"log"
	processor "test"
)

var graphql *string
var rabbit *string

func init() {
	graphql = flag.String("graphql", "localhost/v1/example", "endpoint for graphql")
	rabbit = flag.String("rabbit", "localhost/example", "rabbitMQ endpoint")
}

func main() {
	flag.Parse()

	p, err := processor.NewProcessor(context.Background(), *rabbit, "txs", *graphql)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(p.Process(context.Background()))
}
