package main

import (
	"context"
	"log"
	"os"

	graphqlAPI "github.com/machinebox/graphql"
	watcher "github.com/mapofzones/cosmos-watcher/types"
	processor "github.com/mapofzones/txs-processor/pkg"
	"github.com/mapofzones/txs-processor/pkg/graphql"
	"github.com/mapofzones/txs-processor/pkg/types"
	"github.com/tendermint/tendermint/rpc/client/http"
)

func main() {
	p := processor.Processor{
		Blocks:        mockWatcher(os.Getenv("TMRPC")),
		GraphqlClient: graphqlAPI.NewClient(os.Getenv("GRAPHQL")),
	}

	log.Fatal(p.Process(context.Background()))
}

func mockWatcher(endpoint string) <-chan types.Block {
	tm, err := http.New(endpoint, "/websocket")
	if err != nil {
		panic(err)
	}

	blocks := make(chan types.Block)

	go func() {
		defer func() {
			close(blocks)
			if r := recover(); r != nil {
				log.Println(r)
			}

		}()

		for {

			info, err := tm.Status()
			if err != nil {
				panic(err)
			}

			lastBlock, err := graphql.LastProcessedBlock(context.Background(), info.NodeInfo.Network)
			if err != nil {
				panic(err)
			}

			N := lastBlock + 1

			if N > info.SyncInfo.LatestBlockHeight {
				return
			}

			for N < info.SyncInfo.LatestBlockHeight {
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
		}
	}()

	return blocks
}
