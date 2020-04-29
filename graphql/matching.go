package graphql

import (
	"context"

	"github.com/mapofzones/txs-processor/types"
	"github.com/shurcooL/graphql"
)

// GetUnmatchedIbcTransfers returns slice of txs which are ibc and currently are unmatched
func (c *Client) GetUnmatchedIbcTransfers(ctx context.Context, limit, offset int) ([]types.Transfer, error) {
	var query struct {
		Transfers []struct {
			Hash      string    `graphql:"hash"`
			Quantity  int64     `graphql:"quantity"`
			Recipient string    `graphql:"recipient"`
			Sender    string    `graphql:"sender"`
			T         timestamp `graphql:"timestamp"`
			Token     string    `graphql:"token"`
			Type      string    `graphql:"type"`
			Zone      string    `graphql:"zone"`
		} `graphql:"ibc_tx_transfer_log(where: {matched_to_tx: {_eq: false}}, limit: $limit, offset: $offset)"`
	}

	variables := map[string]interface{}{
		"limit":  graphql.Int(limit),
		"offset": graphql.Int(offset),
	}

	err := c.c.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	// improve error-handling
	if len(query.Transfers) == 0 {
		return []types.Transfer{}, nil
	}

	txs := make([]types.Transfer, len(query.Transfers))
	for i, transfer := range query.Transfers {
		txs[i] = types.Transfer{
			Quantity:  transfer.Quantity,
			Recipient: transfer.Recipient,
			Sender:    transfer.Sender,
			Timestamp: types.Timestamp(transfer.T),
			Token:     transfer.Token,
			Type:      transfer.Type,
			Zone:      transfer.Zone,
			Hash:      transfer.Hash,
		}
	}

	return txs, nil
}

type MatchData struct {
	Hash string
	Zone string
}

// FindMatch returns hash if tx is matched, empty string otherwise
func (c *Client) FindMatch(ctx context.Context, t types.Transfer) (*MatchData, error) {
	var query struct {
		Q []struct {
			Hash string `graphql:"hash"`
			Zone string `graphql:"zone"`
		} `graphql:"ibc_tx_transfer_log(where: {matched_to_tx: {_eq: false}, type: {_eq: $type}, quantity: {_eq: $quantity}, recipient: {_eq: $recipient}, sender: {_eq: $sender}, token: {_eq: $token}, zone: {_neq: $zone}})"`
	}

	variables := map[string]interface{}{
		"type":      graphql.String(types.OppositeType(t.Type)),
		"quantity":  graphql.Int(t.Quantity),
		"recipient": graphql.String(t.Sender),
		"sender":    graphql.String(t.Recipient),
		"token":     graphql.String(t.Token),
		"zone":      graphql.String(t.Zone),
	}

	err := c.c.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	if len(query.Q) > 0 {
		return &MatchData{Hash: query.Q[0].Hash, Zone: query.Q[0].Zone}, nil
	}

	return nil, nil
}

// Match marks transaction as being matched
func (c *Client) Match(ctx context.Context, hash string) error {
	var mutation struct {
		M struct {
			Rows int `graphql:"affected_rows"`
		} `graphql:"update_ibc_tx_transfer_log(where: {hash: {_eq: $hash}}, _set: {matched_to_tx: true})"`
	}

	variables := map[string]interface{}{
		"hash": graphql.String(hash),
	}

	return c.c.Mutate(ctx, &mutation, variables)
}
