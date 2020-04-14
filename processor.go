package processor

import (
	"context"
	"fmt"
	"log"
	"test/graphql"
	rabbit "test/rabbit_mq"
	"time"

	types "test/types"
)

// Processor holds handles for all our connections
type Processor struct {
	cl *graphql.Client
	// ch  *amqp.Channel
	txs <-chan types.Txs
}

// NewProcessor returns instande of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, amqpEndpoint, queueName, graphqlEndopoint string) (*Processor, error) {
	txs, err := rabbit.TxQueue(ctx, amqpEndpoint, queueName)
	if err != nil {
		return nil, err
	}

	client := graphql.NewClient(graphqlEndopoint)
	return &Processor{
		cl:  client,
		txs: txs,
	}, nil
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// launch matcher each n seconds
	go func() {
		for range time.After(time.Second * 5) {
			err := p.processIbc(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	// recieve and send txs
	for {
		select {
		case data := <-p.txs:
			err := p.sendData(ctx, data)
			if err != nil {
				// now we just log
				log.Println(err)
			}
			// we should also ack here after
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Processor) sendData(ctx context.Context, txs types.Txs) error {
	if len(txs.Txs) == 0 {
		return nil
	}
	// first update total tx stats
	// at this moment it also creates zones with all those transactions
	err := p.updateTxStats(ctx, txs.ToStats())
	if err != nil {
		return err
	}

	// insert all the transactions into tx_log table
	rows, err := p.cl.InsertTxs(ctx, txs.Txs)
	if err != nil {
		return err
	}
	if rows != len(txs.Txs) {
		return fmt.Errorf("expected to inser %d transactions into tx_log table, inserted %d", len(txs.Txs), rows)
	}
	return nil
}

func (p *Processor) updateTxStats(ctx context.Context, stats []types.TxStats) error {
	// we only need to check for one object because one message from amqp contains txs from one chain
	exists, err := p.cl.ZoneExists(ctx, stats[0].Zone)
	if err != nil {
		return err
	}

	if !exists {
		err = p.cl.AddZone(ctx, stats[0].Zone, true)
		if err != nil {
			return err
		}
	}

	for _, s := range stats {
		err = p.cl.TotalTxUpsert(ctx, s)
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

		txSlice, err := p.cl.GetUnmatchedIbcTxs(ctx, limit, offset)
		if err != nil {
			return err
		}
		// not txs to process
		if len(txSlice) == 0 {
			return nil
		}
		// map for fast search
		txs := txsHashMap(txSlice)
		// ibc stats map
		stats := types.IbcData{}

		for _, tx := range txs {
			result, err := p.cl.FindMatch(ctx, tx.Recipient, tx.Sender, tx.Quantity, tx.Denom, opositeType(string(tx.Type)))
			if err != nil {
				return err
			}
			// we have matched txs
			if result.Match {
				// delete it from map so we don't have to match it again
				delete(txs, result.Hash)
				// if tx is source
				if tx.Type == types.IbcSend {
					txZone, err := p.cl.ZoneName(ctx, tx.Network)
					if err != nil {
						return err
					}
					stats.Append(txZone, result.Zone, tx.T)
				}
				// if the tx we got is ibc destination
				if tx.Type == types.IbcRecieve {
					txZone, err := p.cl.ZoneName(ctx, tx.Network)
					if err != nil {
						return err
					}
					stats.Append(result.Zone, txZone, tx.T)
				}
				//match them in tx log table
				err = p.cl.MatchTx(ctx, tx.Hash)
				if err != nil {
					return err
				}
				err = p.cl.MatchTx(ctx, result.Hash)
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

func opositeType(s string) string {
	if s == string(types.IbcSend) {
		return string(types.IbcRecieve)
	}
	return string(types.IbcSend)
}

// we use this map to not double process txs if we matched it already and it is in our slice
func txsHashMap(txs []types.Tx) map[string]types.Tx {
	m := map[string]types.Tx{}

	for _, tx := range txs {
		m[tx.Hash] = tx
	}

	return m
}
