package processor

import (
	"context"
	"os"
	"testing"

	graphqlClient "github.com/machinebox/graphql"
	watcher "github.com/mapofzones/cosmos-watcher/types"
	types "github.com/mapofzones/txs-processor/pkg/types"
	"github.com/tendermint/tendermint/rpc/client/http"
)

func TestHub(t *testing.T) {
	p := Processor{
		graphqlClient: graphqlClient.NewClient(os.Getenv("GRAPHQL")),
		blocks:        hubBlocks(),
	}
	t.Fatal(p.Process(context.Background()))
}

func hubBlocks() <-chan types.Block {
	tm, err := http.New("tcp://35.233.155.199:26657", "/websocket")
	if err != nil {
		panic(err)
	}

	blocks := make(chan types.Block)

	go func() {
		N := int64(3680)
		for {
			block, err := tm.Block(&N)
			if err != nil {
				panic(err)
			}

			s := []watcher.TxStatus{}
			for _, tx := range block.Block.Txs {
				res, err := tm.Tx(tx.Hash(), false)
				if err != nil {
					panic(err)
				}
				s = append(s, watcher.TxStatus{
					ResultCode: res.TxResult.Code,
					Hash:       tx.Hash(),
					Height:     res.Height,
				})
			}

			blocks <- types.Block{
				ChainID: block.Block.ChainID,
				Height:  block.Block.Height,
				T:       block.Block.Time,
				Txs:     block.Block.Txs,
				Results: s,
			}
			N++
		}
	}()

	return blocks
}
