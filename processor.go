package processor

import (
	"context"
	"log"
	"time"

	"errors"

	"github.com/mapofzones/txs-processor/graphql"
	rabbit "github.com/mapofzones/txs-processor/rabbit_mq"

	types "github.com/mapofzones/txs-processor/types"
)

// Processor holds handles for all our connections
type Processor struct {
	cl *graphql.Client
	// ch  *amqp.Channel
	blocks <-chan types.Block
}

// NewProcessor returns instance of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, amqpEndpoint, queueName, graphqlEndpoint string) (*Processor, error) {
	txs, err := rabbit.BlockStream(ctx, amqpEndpoint, queueName)
	if err != nil {
		return nil, err
	}

	client := graphql.NewClient(graphqlEndpoint)
	return &Processor{
		cl:     client,
		blocks: txs,
	}, nil
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// launch matcher each n seconds
	go func() {
		for range time.After(time.Minute) {
			err := p.processIbc(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}()

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
	validTxs, err := block.GetValidStdTxs()
	if err != nil {
		return err
	}

	// at this moment it also creates zones with all those transactions
	// this is temporary before finer-grade control over zones is used
	zoneName, err := p.cl.ZoneName(ctx, block.ChainID)
	if err != nil {
		return err
	}
	if zoneName == "" {
		zoneName = block.ChainID
		if err := p.addZone(ctx, block.ChainID); err != nil {
			return err
		}
	}

	// update total_tx_stats
	if len(validTxs) > 0 {
		err = p.updateTxStats(ctx, []types.TxStats{block.ToTxStats(zoneName, len(validTxs))})
		if err != nil {
			return err
		}
	}

	// check transactions for having ibc messages inside them
	for _, tx := range validTxs {
		err := p.cl.ProcessIbcTx(ctx, zoneName, tx, block.T)
		if err != nil {
			return err
		}
	}

	// tell system that the block is processed
	err = p.cl.AddBlock(ctx, block)

	return err
}

func (p *Processor) addZone(ctx context.Context, chainID string) error {
	// we only need to check for one object because one message from amqp contains txs from one chain
	exists, err := p.cl.ZoneExists(ctx, chainID)
	if err != nil {
		return err
	}

	if !exists {
		err = p.cl.AddZone(ctx, chainID, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) updateTxStats(ctx context.Context, stats []types.TxStats) error {
	for _, s := range stats {
		err := p.cl.TotalTxUpsert(ctx, s)
		if err != nil {
			return err
		}
	}
	return nil
}

// processIbc iterates through unmatched ibc txs and matches them
func (p *Processor) processIbc(ctx context.Context) error {
	// maybe should be constants
	limit := 100
	offset := 0
	// loop, exit if we have got 0 tex
	for {

		transfers, err := p.cl.GetUnmatchedIbcTransfers(ctx, limit, offset)
		if err != nil {
			return err
		}
		// not txs to process
		if len(transfers) == 0 {
			return nil
		}
		// map for fast search
		txs := txsHashMap(transfers)
		// ibc stats map
		stats := types.IbcData{}

		for _, transfer := range txs {
			Match, err := p.cl.FindMatch(ctx, transfer)
			if err != nil {
				return err
			}

			// we have matched transfers
			if Match != nil {
				// delete it from map so we don't have to match it again
				delete(txs, Match.Hash)
				// if tx is source
				if transfer.Type == types.Send {
					stats.Append(transfer.Zone, Match.Zone, types.FromTimestamp(transfer.Timestamp))
				}
				// if the tx we got is ibc destination
				if transfer.Type == types.Receive {
					stats.Append(Match.Zone, transfer.Zone, types.FromTimestamp(transfer.Timestamp))
				}
				//match them in tx log table
				err = p.cl.Match(ctx, transfer.Hash)
				if err != nil {
					return err
				}
				err = p.cl.Match(ctx, Match.Hash)
				if err != nil {
					return err
				}
			}
		}
		err = p.cl.IbcTxUpsert(ctx, stats.ToIbcStats()...)
		if err != nil {
			return err
		}
	}
}

// we use this map to not double process transfers if we matched it already and it is in our slice
func txsHashMap(txs []types.Transfer) map[string]types.Transfer {
	m := map[string]types.Transfer{}

	for _, transfer := range txs {
		m[transfer.Hash] = transfer
	}

	return m
}
