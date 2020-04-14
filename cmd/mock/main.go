package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	processor "test/types"
	"time"
)

func main() {
	txs := []processor.Tx{}

	for i := 0; i < 1000; i++ {
		txs = append(txs, RandomTx())
	}
	bytes, _ := json.MarshalIndent(txs, "", "  ")

	ioutil.WriteFile("txs.json", bytes, 0644)
}

func RandomTx() processor.Tx {
	names := []string{"Alice", "Bob", "Jack"}
	amounts := []string{"1", "2", "3"}
	denoms := []string{"copper", "gold"}
	zones := []string{"BTC", "eth"}
	types := []processor.Type{processor.IbcSend, processor.IbcRecieve, processor.Transfer}

	return processor.Tx{
		T:         time.Now(),
		Denom:     denoms[rand.Intn(len(denoms))],
		Quantity:  amounts[rand.Intn(len(amounts))],
		Network:   zones[rand.Intn(len(zones))],
		Sender:    names[rand.Intn(len(names))],
		Recipient: names[rand.Intn(len(names))],
		Type:      types[rand.Intn(len(types))],
		Hash:      bytes(),
	}
}

func bytes() string {
	bytes := make([]byte, 100)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}
