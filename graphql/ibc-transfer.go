package graphql

import (
	"context"

	"github.com/mapofzones/txs-processor/types"
	"github.com/shurcooL/graphql"
)

// alias we need for queries
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
