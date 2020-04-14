package processor

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"test/graphql"
	processor "test/types"
	"testing"

	"github.com/attractor-spectrum/cosmos-watcher/tx"
)

const endpoint = ""

func TestProcessor(t *testing.T) {
	bytes, err := ioutil.ReadFile("txs.json")
	if err != nil {
		t.Fatal(err)
	}

	txs := []processor.Tx{}
	err = json.Unmarshal(bytes, &txs)
	if err != nil {
		t.Fatal(err)
	}
	txsChan := func() <-chan processor.Txs {
		txsChan := make(chan processor.Txs)
		go func() {
			for i := 0; i < len(txs)/5; i++ {
				txsChan <- processor.Txs{Txs: []tx.Tx{txs[i]}}
			}
		}()
		return txsChan
	}

	p := &Processor{
		txs: txsChan(),
		cl:  graphql.NewClient(endpoint),
	}

	err = p.Process(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
