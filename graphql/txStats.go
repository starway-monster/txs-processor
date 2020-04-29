package graphql

import (
	"context"
	"errors"

	types "github.com/mapofzones/txs-processor/types"

	"github.com/shurcooL/graphql"
)

// totalTxExists returns true if column already exists in db, else false
func (c *Client) totalTxExists(ctx context.Context, stats types.TxStats) (bool, error) {
	data, err := c.totalTxGet(ctx, stats)
	if err != nil {
		return false, err
	}

	return (data.Count > 0), nil
}

// totalTxGet returns TxStats struct, which represents one line in database, bool represents if a table with this record already exists
func (c *Client) totalTxGet(ctx context.Context, stats types.TxStats) (types.TxStats, error) {
	var query struct {
		Stats []struct {
			Zone  graphql.String `graphql:"zone"`
			Hour  timestamp      `graphql:"hour"`
			Count graphql.Int    `graphql:"txs_cnt"`
		} `graphql:" total_tx_hourly_stats(where: {zone: {_eq: $zone}, hour: {_eq: $hour}})"`
	}

	variables := map[string]interface{}{
		"zone": graphql.String(stats.Zone),
		"hour": timestamp(stats.Hour.Format(types.Format)),
	}
	if err := c.c.Query(ctx, &query, variables); err != nil {
		return types.TxStats{}, err
	}
	if len(query.Stats) < 1 {
		return types.TxStats{}, nil
	}

	return types.TxStats{
		Zone:  string(query.Stats[0].Zone),
		Count: int(query.Stats[0].Count),
		Hour:  types.FromTimestamp(types.Timestamp(query.Stats[0].Hour)),
	}, nil
}

func (c *Client) totalTxAdd(ctx context.Context, stats types.TxStats) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"insert_total_tx_hourly_stats(objects: {hour: $hour, txs_cnt: $txs, zone: $zone})"`
	}
	variables := map[string]interface{}{
		"zone": graphql.String(stats.Zone),
		"hour": timestamp(stats.Hour.Format(types.Format)),
		"txs":  graphql.Int(stats.Count),
	}

	err := c.c.Mutate(ctx, &mutation, variables)
	if err == nil && mutation.M.Rows != 1 {
		return errors.New("invalid amount of rows edited")
	}
	return err
}

func (c *Client) totalTxIncrement(ctx context.Context, stats types.TxStats) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"update_total_tx_hourly_stats(where: {hour: {_eq: $hour}, zone: {_eq: $zone}}, _inc: {txs_cnt: $txs})"`
	}
	variables := map[string]interface{}{
		"zone": graphql.String(stats.Zone),
		"hour": timestamp(stats.Hour.Format(types.Format)),
		"txs":  graphql.Int(stats.Count),
	}

	err := c.c.Mutate(ctx, &mutation, variables)
	if err == nil && mutation.M.Rows != 1 {
		return errors.New("invalid amount of rows edited")
	}

	return err
}

// TotalTxUpsert creates new row if it doesn't exist or updates it if it exists
func (c *Client) TotalTxUpsert(ctx context.Context, stats types.TxStats) error {
	exists, err := c.totalTxExists(ctx, stats)
	if err != nil {
		return err
	}

	if !exists {
		return c.totalTxAdd(ctx, stats)

	}
	return c.totalTxIncrement(ctx, stats)
}
