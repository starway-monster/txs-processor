package processor

import (
	"fmt"
	"testing"
	"time"

	watcher "github.com/attractor-spectrum/cosmos-watcher"
	"github.com/attractor-spectrum/cosmos-watcher/tx"
)

func TestSplitTime(t *testing.T) {
	data := Txs{Err: nil, Txs: []tx.Tx{
		Tx{
			T:       time.Now(),
			Network: "A",
		},
		Tx{
			T:       time.Now(),
			Network: "A",
		},
		tx.Tx{
			T:       time.Now().AddDate(0, 0, -1),
			Network: "A",
		},
		tx.Tx{
			T:       time.Now(),
			Network: "B",
		},
	}}

	fmt.Println(data.ToStats())
}

func TestSplitIBC(t *testing.T) {
	data := Txs{Err: nil, Txs: []tx.Tx{
		Tx{
			Type:    watcher.IbcSend,
			Network: "A",
		},
		Tx{
			Type:    watcher.IbcRecieve,
			Network: "A",
		},
		tx.Tx{
			Type:    watcher.Transfer,
			Network: "A",
		},
		tx.Tx{
			Type:    watcher.Other,
			Network: "B",
		},
	}}

	local, ibc := data.SplitIBC()

	for _, tx := range local.Txs {
		if tx.Type == watcher.IbcRecieve || tx.Type == watcher.IbcSend {
			t.Fatalf("expected non-ibc trasnactions, got %v", tx)
		}
	}

	for _, tx := range ibc.Txs {
		if tx.Type != watcher.IbcRecieve && tx.Type != watcher.IbcSend {
			t.Fatalf("expected ibc transactions, got %v", tx)
		}
	}
}
