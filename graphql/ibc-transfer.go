package graphql

import (
	"context"
	"errors"
	"time"

	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	"github.com/mapofzones/txs-processor/types"
	"github.com/shurcooL/graphql"
)

func (c *Client) ProcessIbcTx(ctx context.Context, zone string, tx types.TxWithHash, blockTime time.Time) error {
	for _, msg := range tx.Tx.Msgs {
		// if we have transfer, i.e. blockchain is sending to another chain
		if msg.Type() == "transfer" {
			if transfer, ok := msg.(transfer.MsgTransfer); ok {
				if err := c.sendIbcTransfer(ctx, types.FromMsgTransfer(transfer, tx.Hash, zone, blockTime)); err != nil {
					return err
				}
			} else {
				return errors.New("could not cast interface to type MsgTransfer, probably invalid cosmos sdk version")
			}
		}
		// if we are receiving packet
		if msg.Type() == "ics04/opaque" {
			if packetMsg, ok := msg.(channel.MsgPacket); ok {
				var data transfer.FungibleTokenPacketData
				if err := types.Codec.UnmarshalJSON(packetMsg.Packet.GetData(), &data); err != nil {
					return errors.New("could not unmarshal packet data into FungibleTOkenPacketData")
				}
				err := c.sendIbcTransfer(ctx, types.FromMsgPacket(data, tx.Hash, zone, blockTime))
				if err != nil {
					return err
				}
			} else {
				return errors.New("could not cast interface to type MsgPacket, probably invalid cosmos sdk version")
			}
		}
	}
	return nil
}

type timestamp types.Timestamp

func (c *Client) sendIbcTransfer(ctx context.Context, t types.Transfer) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"insert_ibc_tx_transfer_log(objects: {hash: $hash, quantity: $quantity, recipient: $recipient, sender: $sender, timestamp: $timestamp, token: $token, type: $type, zone: $zone})"`
	}

	variables := map[string]interface{}{
		"hash":      graphql.String(t.Hash),
		"quantity":  graphql.Int(t.Quantity),
		"sender":    graphql.String(t.Sender),
		"recipient": graphql.String(t.Recipient),
		"timestamp": timestamp(t.Timestamp),
		"token":     graphql.String(t.Token),
		"type":      graphql.String(t.Type),
		"zone":      graphql.String(t.Zone),
	}

	return c.c.Mutate(ctx, &mutation, variables)

}
