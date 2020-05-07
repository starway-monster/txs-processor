package processor

import (
	"context"
	"fmt"
	"log"

	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mutationClient "github.com/machinebox/graphql"
	"github.com/mapofzones/txs-processor/pkg/builder"
	"github.com/mapofzones/txs-processor/pkg/graphql"
	rabbit "github.com/mapofzones/txs-processor/pkg/rabbit_mq"
	types "github.com/mapofzones/txs-processor/pkg/types"
)

// Processor holds handles for all our connections
type Processor struct {
	graphqlClient *mutationClient.Client
	blocks        <-chan types.Block
}

// NewProcessor returns instance of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, amqpEndpoint, queueName, graphqlEndpoint string) (*Processor, error) {
	txs, err := rabbit.BlockStream(ctx, amqpEndpoint, queueName)
	if err != nil {
		return nil, err
	}

	return &Processor{
		blocks:        txs,
		graphqlClient: mutationClient.NewClient(graphqlEndpoint),
	}, nil
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// receive and send txs
	for {
		select {
		case data, ok := <-p.blocks:
			if !ok {
				return errors.New("block channel is closed")
			}
			err := p.sendData(ctx, data)
			if err != nil {
				// now we just log
				log.Println(err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Processor) sendData(ctx context.Context, block types.Block) error {
	builder := builder.MutationBuilder{}

	// check if we got the block we needed
	height, err := graphql.LastProcessedBlock(ctx, block.ChainID)
	if err != nil {
		return err
	}
	if height+1 != block.Height {
		return fmt.Errorf("expected to get block at height: %d, got at: %d", height+1, block.Height)
	}

	builder.AddZone(block.ChainID)

	builder.MarkBlock(block)

	// // update total_tx_stats
	// if len(validTxs) > 0 {
	// 	txStats := block.ToTxStats(zoneName, len(validTxs))
	// 	exists, err := p.queryClient.TotalTxExists(ctx, txStats)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if exists {
	// 		builder.UpdateTxStats(txStats)
	// 	} else {
	// 		builder.CreateTxStats(txStats)
	// 	}
	// }

	err = processBlock(ctx, block, &builder)
	if err != nil {
		return err
	}
	fmt.Println(builder.Mutation())
	return p.graphqlClient.Run(ctx, mutationClient.NewRequest(builder.Mutation()), nil)
}

func processBlock(ctx context.Context, block types.Block, b *builder.MutationBuilder) error {
	// get all successful transactions
	validTxs, err := block.GetValidStdTxs()
	if err != nil {
		log.Println("could not decode tx from ", block.ChainID)
		return err
	}
	// here we should also submit total tx count,but we currently don't

	// process messages
	msgs := []sdk.Msg{}
	for _, tx := range validTxs {
		msgs = append(msgs, tx.Tx.Msgs...)
	}

	return processMsgs(ctx, block, b, msgs)
}
