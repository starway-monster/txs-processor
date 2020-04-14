package processor

import (
	"fmt"
	"testing"
	"time"
)

func TestSplitTime(t *testing.T) {
	data := Txs{Err: nil, Txs: []Tx{
		Tx{
			T:       time.Now(),
			Network: "A",
		},
		Tx{
			T:       time.Now(),
			Network: "A",
		},
		Tx{
			T:       time.Now().AddDate(0, 0, -1),
			Network: "A",
		},
		Tx{
			T:       time.Now(),
			Network: "B",
		},
	}}

	fmt.Println(data.ToStats())
}

func TestSplitIBC(t *testing.T) {
	data := Txs{Err: nil, Txs: []Tx{
		Tx{
			Type:    IbcSend,
			Network: "A",
		},
		Tx{
			Type:    IbcRecieve,
			Network: "A",
		},
		Tx{
			Type:    Transfer,
			Network: "A",
		},
		Tx{
			Type:    Other,
			Network: "B",
		},
	}}

	local, ibc := data.SplitIBC()

	for _, tx := range local.Txs {
		if tx.Type == IbcRecieve || tx.Type == IbcSend {
			t.Fatalf("expected non-ibc trasnactions, got %v", tx)
		}
	}

	for _, tx := range ibc.Txs {
		if tx.Type != IbcRecieve && tx.Type != IbcSend {
			t.Fatalf("expected ibc transactions, got %v", tx)
		}
	}
}
