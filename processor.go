package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"errors"

	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	mutationClient "github.com/machinebox/graphql"
	"github.com/mapofzones/txs-processor/builder"
	"github.com/mapofzones/txs-processor/graphql"
	rabbit "github.com/mapofzones/txs-processor/rabbit_mq"
	types "github.com/mapofzones/txs-processor/types"
)

// Processor holds handles for all our connections
type Processor struct {
	queryClient *graphql.Client
	// ch  *amqp.Channel
	mutationClient *mutationClient.Client
	blocks         <-chan types.Block
}

// NewProcessor returns instance of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, amqpEndpoint, queueName, graphqlEndpoint string) (*Processor, error) {
	txs, err := rabbit.BlockStream(ctx, amqpEndpoint, queueName)
	if err != nil {
		return nil, err
	}

	client := graphql.NewClient(graphqlEndpoint)
	return &Processor{
		queryClient:    client,
		blocks:         txs,
		mutationClient: mutationClient.NewClient(graphqlEndpoint),
	}, nil
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// launch matcher each n seconds
	go func() {
		for range time.Tick(time.Minute * 1) {
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
	builder := builder.MutationBuilder{}

	validTxs, err := block.GetValidStdTxs()
	if err != nil {
		fmt.Println("could not decode tx from ", block.ChainID)
		return err
	}

	// at this moment it also creates zones with all those transactions
	// this is temporary before finer-grade control over zones is used
	zoneName, err := p.queryClient.ZoneName(ctx, block.ChainID)
	if err != nil {
		return err
	}
	if zoneName == "" {
		zoneName = block.ChainID
		builder.AddZone(zoneName)
	}

	builder.PushBlock(block)

	// update total_tx_stats
	if len(validTxs) > 0 {
		txStats := block.ToTxStats(zoneName, len(validTxs))
		exists, err := p.queryClient.TotalTxExists(ctx, txStats)
		if err != nil {
			return err
		}
		if exists {
			builder.UpdateTxStats(txStats)
		} else {
			builder.CreateTxStats(txStats)
		}
	}

	// check transactions for having ibc messages inside them
	for _, tx := range validTxs {
		transfers, err := p.processIbcTx(zoneName, tx, block.T)
		if err != nil {
			return err
		}
		for _, transfer := range transfers {
			builder.PushTransfer(transfer)
		}
	}

	return p.mutationClient.Run(ctx, mutationClient.NewRequest(builder.Mutation()), nil)
}

// processIbcTx returns set of transfers from a transaction
func (c *Processor) processIbcTx(zone string, tx types.TxWithHash, blockTime time.Time) ([]types.Transfer, error) {
	transfers := []types.Transfer{}
	for _, msg := range tx.Tx.Msgs {
		// if we have transfer, i.e. blockchain is sending to another chain
		if msg.Type() == "transfer" {
			if transfer, ok := msg.(transfer.MsgTransfer); ok {
				transfers = append(transfers, types.FromMsgTransfer(transfer, tx.Hash, zone, blockTime))
			} else {
				return nil, errors.New("could not cast interface to type MsgTransfer, probably invalid cosmos sdk version")
			}
		}
		// if we are receiving packet
		if msg.Type() == "ics04/opaque" {
			if packetMsg, ok := msg.(channel.MsgPacket); ok {
				var data transfer.FungibleTokenPacketData
				if err := types.Codec.UnmarshalJSON(packetMsg.Packet.GetData(), &data); err != nil {
					return nil, errors.New("could not unmarshal packet data into FungibleTokenPacketData")
				}
				transfers = append(transfers, types.FromMsgPacket(data, tx.Hash, zone, blockTime))
			} else {
				return nil, errors.New("could not cast interface to type MsgPacket, probably invalid cosmos sdk version")
			}
		}
	}
	return transfers, nil
}

// processIbc iterates through unmatched ibc txs and matches them
func (p *Processor) processIbc(ctx context.Context) error {
	// maybe should be constants
	limit := 10000
	offset := 0
	// loop, exit if we have got 0 transfers
	for ; ; offset += limit {

		transfers, err := p.queryClient.GetUnmatchedIbcTransfers(ctx, limit, offset)
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
			Match, err := p.queryClient.FindMatch(ctx, transfer)
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
				err = p.queryClient.Match(ctx, transfer.Hash)
				if err != nil {
					return err
				}
				err = p.queryClient.Match(ctx, Match.Hash)
				if err != nil {
					return err
				}
			}
		}
		err = p.queryClient.IbcTxUpsert(ctx, stats.ToIbcStats()...)
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
