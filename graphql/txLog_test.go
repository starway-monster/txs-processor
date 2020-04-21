package graphql

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	types "test/types"
	"testing"
	"time"
)

func randBytes() []byte {
	b := make([]byte, 10)
	rand.Read(b)
	return b
}

func TestInsertTx(t *testing.T) {
	tx := types.Tx{
		T:         time.Now(),
		Denom:     "token",
		Hash:      base64.RawStdEncoding.EncodeToString(randBytes()),
		Network:   "z10018",
		Precision: 0,
		Quantity:  "1",
		Sender:    "Robin Hood",
		Recipient: "The poor",
		Type:      types.Transfer,
		Data:      nil,
	}

	c := NewClient(endpoint)
	rows, err := c.InsertTx(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	if rows != 1 {
		t.Fatalf("expected to inser one row, inserted %d", rows)
	}
}

func TestInsertTxs(t *testing.T) {
	txA := types.Tx{
		T:         time.Now(),
		Denom:     "token",
		Hash:      base64.RawStdEncoding.EncodeToString(randBytes()),
		Network:   "z10018",
		Precision: 0,
		Quantity:  "1",
		Sender:    "Robin Hood",
		Recipient: "The poor",
		Type:      types.Transfer,
		Data:      nil,
	}

	txB := types.Tx{
		T:         time.Now().AddDate(0, -1, 0),
		Denom:     "not_token",
		Hash:      base64.RawStdEncoding.EncodeToString(randBytes()),
		Network:   "z279",
		Precision: 0,
		Quantity:  "100.000",
		Sender:    "Bob",
		Recipient: "Alice",
		Type:      types.IbcSend,
		Data:      nil,
	}

	c := NewClient(endpoint)
	rows, err := c.InsertTxs(context.Background(), []types.Tx{txA, txB})
	if err != nil {
		t.Fatal(err)
	}

	if rows != 2 {
		t.Fatalf("expected to inser one row, inserted %d", rows)
	}
}

func TestFindTxs(t *testing.T) {
	c := NewClient(endpoint)

	hashes, err := c.FindTxs(context.Background(), "Bob", "Alice", "100.000", "not_token", "ibc-send", true, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(hashes) == 0 {
		t.Fatal("did not recieve hash for tx")
	}

	for _, hash := range hashes {
		if hash == "r3+pcqtAUfjLIg" {
			return // we got our hash, stuff is working
		}
	}
	t.Fatal("did not recieve valid hash")
}

func TestTxByType(t *testing.T) {
	c := NewClient(endpoint)

	_, err := c.TxByHash(context.Background(), "r3+pcqtAUfjLIg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMatchTx(t *testing.T) {
	c := NewClient(endpoint)

	err := c.MatchTx(context.Background(), "G00HOnc55xYzXw")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmatchedIbcTx(t *testing.T) {
	c := NewClient(endpoint)

	txs, err := c.GetUnmatchedIbcTxs(context.Background(), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(txs)
}

func TestFindMatch(t *testing.T) {
	c := NewClient(endpoint)

	hash, err := c.FindMatch(context.Background(), "Bob", "Alice", "1", "copper", "ibc-send")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(hash)
}

func TestUnmatchedIbcSend(t *testing.T) {
	c := NewClient(endpoint)

	txs, err := c.GetUnmatchedIbcSend(context.Background(), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(txs)
}

func TestFindSendMatch(t *testing.T) {
	c := NewClient(endpoint)

	hash, err := c.FindSendMatch(context.Background(), types.Tx{Sender: "cosmos1ewzutwrj9vvvr8qzanq24959zg9mmj6pg2022d",
		Recipient: "cosmos1jenzfq8wjexx5e638jq803v76j6y9pzcyd8592",
		Network:   "irishub",
		Quantity:  "100000",
		Denom:     "okt"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(hash)
}
