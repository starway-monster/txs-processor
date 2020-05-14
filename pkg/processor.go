package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rabbit "github.com/mapofzones/txs-processor/pkg/rabbit_mq"
	types "github.com/mapofzones/txs-processor/pkg/types"
	processor "github.com/mapofzones/txs-processor/pkg/x"
)

// Processor holds handles for all our connections
// and holds an interface which defines what has to be done
// on the received block
type Processor struct {
	Blocks      <-chan types.Block
	impl        processor.Processor
	heightCache map[string]int64
}

// NewProcessor returns instance of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, amqpEndpoint, queueName, p processor.Processor) (*Processor, error) {
	txs, err := rabbit.BlockStream(ctx, amqpEndpoint, queueName)
	if err != nil {
		return nil, err
	}

	return &Processor{
		Blocks:      txs,
		impl:        p,
		heightCache: map[string]int64{},
	}, nil
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// receive and send txs
	for {
		select {
		case data, ok := <-p.Blocks:
			if !ok {
				return errors.New("block channel is closed")
			}
			err := p.sendData(ctx, data)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Processor) sendData(ctx context.Context, block types.Block) error {
	p.impl.Reset()

	p.impl.AddZone(block.ChainID)
	// check if we got the block at the
	// required height
	var height int64

	if cachedHeight, ok := p.heightCache[block.ChainID]; ok {
		height = cachedHeight
	} else {

		h, err := p.impl.LastProcessedBlock(block.ChainID)
		if err != nil {
			return err
		}
		height = h
		p.heightCache[block.ChainID] = height
	}

	if height+1 != block.Height {
		return fmt.Errorf("expected to get block at height: %d, got at: %d", height+1, block.Height)
	}

	p.impl.MarkBlock()

	if err = processBlock(ctx, block, p.impl); err != nil {
		return err
	}

	if err = p.impl.Commit(ctx); err != nil {
		return err
	}

	// increment the height at which we expect to get the block
	p.heightCache[block.ChainID]++
	return nil
}

func processBlock(ctx context.Context, block types.Block, p processor.Processor) error {
	// get all successful transactions
	validTxs, err := block.GetValidStdTxs()
	if err != nil {
		log.Println("could not decode tx from ", block.ChainID)
		return err
	}

	p.AddTxStats(types.TxStats{
		ChainID: block.ChainID,
		Hour:    block.T.Truncate(time.Hour),
		Count:   len(validTxs),
	})

	// process messages
	msgs := make([]sdk.Msg, 0, 300)
	for _, tx := range validTxs {
		msgs = append(msgs, tx.Tx.Msgs...)
	}

	err := processMsgs(ctx, block, p, msgs)
	if err != nil {
		return err
	}
}
