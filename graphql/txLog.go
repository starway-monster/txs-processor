package graphql

import (
	"context"
	types "test/types"

	"github.com/shurcooL/graphql"
)

// InsertTx puts tx to tx_log table for further matching by tx processor
func (c *Client) InsertTx(ctx context.Context, tx types.Tx) (int, error) {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"insert_txs_log(objects: {hash: $hash, is_ibc: $ibc, quantity: $quantity, sender: $sender, recipient: $recipient, time: $time, token: $token, type: $type, zone: $zone})"`
	}

	name, err := c.ZoneName(ctx, tx.Network)
	if err != nil {
		return 0, err
	}

	variables := map[string]interface{}{
		"hash":      graphql.String(tx.Hash),
		"ibc":       graphql.Boolean(types.Ibc(tx)),
		"quantity":  graphql.String(tx.Quantity),
		"sender":    graphql.String(tx.Sender),
		"recipient": graphql.String(tx.Recipient),
		"time":      toTimestamp(tx.T),
		"token":     graphql.String(tx.Denom),
		"type":      graphql.String(tx.Type),
		"zone":      graphql.String(name),
	}

	err = c.c.Mutate(ctx, &mutation, variables)

	if err != nil {
		return 0, err
	}

	return int(mutation.M.Rows), nil
}

// InsertTxs calls Insert tx on each tx object
// might be inefficient but should suffice, the only alternative is building query manualy
func (c *Client) InsertTxs(ctx context.Context, txs []types.Tx) (int, error) {
	count := 0

	for _, tx := range txs {
		rows, err := c.InsertTx(ctx, tx)
		if err != nil {
			return count, err
		}
		count += rows
	}

	return count, nil
}

// TxByHash returns tx with corresponding hash
func (c *Client) TxByHash(ctx context.Context, hash string) (types.Tx, error) {
	var query struct {
		Txs []struct {
			Quantity  string    `graphql:"quantity"`
			Recipient string    `graphql:"recipient"`
			Sender    string    `graphql:"sender"`
			T         timestamp `graphql:"time"`
			Token     string    `graphql:"token"`
			Type      string    `graphql:"type"`
			Zone      string    `graphql:"zone"`
		} `graphql:"txs_log(where: {hash: {_eq: $hash}})"`
	}

	variables := map[string]interface{}{
		"hash": graphql.String(hash),
	}

	err := c.c.Query(ctx, &query, variables)
	if err != nil {
		return types.Tx{}, err
	}

	// improve error-handling
	if len(query.Txs) == 0 {
		return types.Tx{}, nil
	}

	tx := types.Tx{
		Quantity:  query.Txs[0].Quantity,
		Recipient: query.Txs[0].Recipient,
		Sender:    query.Txs[0].Sender,
		T:         fromTimestamp(query.Txs[0].T),
		Denom:     query.Txs[0].Token,
		Type:      types.Type(query.Txs[0].Type),
		Network:   query.Txs[0].Zone,
	}

	return tx, nil
}

// FindTxs finds all matching transactions and returns their hashes
func (c *Client) FindTxs(ctx context.Context, sender, recipient, quantity, token, Type string, ibc, matched bool) ([]string, error) {
	var query struct {
		Hashes []struct {
			Hash string `graphql:"hash"`
		} `graphql:"txs_log(where: {is_ibc: {_eq: $is_ibc}, matched_to_tx: {_eq: $matched}, sender: {_eq: $sender}, recipient: {_eq: $recipient}, quantity: {_eq: $quantity}, token: {_eq: $token}, type: {_eq: $type}})"`
	}

	variables := map[string]interface{}{
		"is_ibc":    graphql.Boolean(ibc),
		"matched":   graphql.Boolean(matched),
		"sender":    graphql.String(sender),
		"recipient": graphql.String(recipient),
		"quantity":  graphql.String(quantity),
		"token":     graphql.String(token),
		"type":      graphql.String(Type),
	}
	err := c.c.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	hashes := []string{}
	for _, i := range query.Hashes {
		hashes = append(hashes, i.Hash)
	}
	return hashes, nil
}

// MatchData is used to couple hash and zone from graphql query
type MatchData struct {
	Hash  string
	Zone  string
	Match bool
}

// FindMatch returns hash if corresponding tx exists,else empty string
func (c *Client) FindMatch(ctx context.Context, sender, recipient, quantity, token, Type string) (MatchData, error) {
	var query struct {
		Hashes []struct {
			Hash string `graphql:"hash"`
			Zone string `graphql:"zone"`
		} `graphql:"txs_log(where: {is_ibc: {_eq: true}, matched_to_tx: {_eq: false}, sender: {_eq: $sender}, recipient: {_eq: $recipient}, quantity: {_eq: $quantity}, token: {_eq: $token}, type: {_eq: $type}})"`
	}

	variables := map[string]interface{}{
		"sender":    graphql.String(sender),
		"recipient": graphql.String(recipient),
		"quantity":  graphql.String(quantity),
		"token":     graphql.String(token),
		"type":      graphql.String(Type),
	}

	err := c.c.Query(ctx, &query, variables)
	if err != nil {
		return MatchData{Match: false}, err
	}

	if len(query.Hashes) == 0 {
		return MatchData{Match: false}, nil
	}
	return MatchData{Hash: query.Hashes[0].Hash, Zone: query.Hashes[0].Zone, Match: true}, nil
}

// MatchTx mark transaction with given cash as matched
func (c *Client) MatchTx(ctx context.Context, hash string) error {
	var mutation struct {
		M struct {
			Rows int `graphql:"affected_rows"`
		} `graphql:"update_txs_log(where: {hash: {_eq: $hash}}, _set: {matched_to_tx: true})"`
	}

	variables := map[string]interface{}{
		"hash": graphql.String(hash),
	}

	err := c.c.Mutate(ctx, &mutation, variables)
	return err
}

// GetUnmatchedIbcTxs returns slice of txs which are ibc and currenty are unmatched
func (c *Client) GetUnmatchedIbcTxs(ctx context.Context, limit, offset int) ([]types.Tx, error) {
	var query struct {
		Txs []struct {
			Quantity  string    `graphql:"quantity"`
			Recipient string    `graphql:"recipient"`
			Sender    string    `graphql:"sender"`
			T         timestamp `graphql:"time"`
			Token     string    `graphql:"token"`
			Type      string    `graphql:"type"`
			Zone      string    `graphql:"zone"`
			Hash      string    `graphql:"hash"`
		} `graphql:"txs_log(limit: $limit, offset: $offset, where: {is_ibc: {_eq: true}, matched_to_tx: {_eq: false}})"`
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
	if len(query.Txs) == 0 {
		return []types.Tx{}, nil
	}

	txs := make([]types.Tx, len(query.Txs))
	for i, tx := range query.Txs {
		txs[i] = types.Tx{
			Quantity:  tx.Quantity,
			Recipient: tx.Recipient,
			Sender:    tx.Sender,
			T:         fromTimestamp(tx.T),
			Denom:     tx.Token,
			Type:      types.Type(tx.Type),
			Network:   tx.Zone,
			Hash:      tx.Hash,
		}
	}

	return txs, nil
}
