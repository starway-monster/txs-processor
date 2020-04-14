package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	r "test/rabbit_mq"
	"time"
)

func main() {
	txs, err := r.TxQueue(context.Background(), "amqp://ggmjxdkq:HJQ4N7gABKrLDWoneYwr0M-qZZDconkO@clam.rmq.cloudamqp.com/ggmjxdkq", "txs")
	if err != nil {
		log.Fatal(err)
	}

	dump := []r.Tx{}

	go func() {

		for txSlice := range txs {
			if txSlice.Err != nil {
				log.Fatal(err)
			}
			for _, tx := range txSlice.Txs {
				if tx.Hash != "" {
					dump = append(dump, tx)
				}
			}
		}
	}()
	time.Sleep(120 * time.Second)
	bytes, _ := json.MarshalIndent(dump, "", "  ")

	ioutil.WriteFile("txs", bytes, 0777)
}
