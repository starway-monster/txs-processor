package graphql

import (
	"context"
	"errors"
	types "test/types"

	"github.com/shurcooL/graphql"
)

// ibcTxExists returns true if column already exists in db, else false
func (c *Client) ibcTxExists(ctx context.Context, stats types.IbcStats) (bool, error) {
	data, err := c.ibcTxGet(ctx, stats)
	if err != nil {
		return false, err
	}

	return (data.Count > 0), nil
}

// ibcTxGet returns IbcStats struct, which represents one line in database, bool represents if a table with this record already exists
func (c *Client) ibcTxGet(ctx context.Context, stats types.IbcStats) (types.IbcStats, error) {
	var query struct {
		Stats []struct {
			Count graphql.Int `graphql:"txs_cnt"`
		} `graphql:"ibc_tx_hourly_stats(where: {zone_dest: {_eq: $dest}, zone_src: {_eq: $src}, hour: {_eq: $hour}})"`
	}

	variables := map[string]interface{}{
		"dest": graphql.String(stats.Destination),
		"src":  graphql.String(stats.Source),
		"hour": timestamp(stats.Hour.Format(types.Format)),
	}
	if err := c.c.Query(ctx, &query, variables); err != nil {
		return types.IbcStats{}, err
	}
	if len(query.Stats) < 1 {
		return types.IbcStats{}, nil
	}

	return types.IbcStats{
		Destination: stats.Destination,
		Count:       int(query.Stats[0].Count),
		Hour:        stats.Hour,
		Source:      stats.Source,
	}, nil
}

func (c *Client) ibcTxAdd(ctx context.Context, stats types.IbcStats) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"insert_ibc_tx_hourly_stats(objects: {hour: $hour, txs_cnt: $count, zone_dest: $dest, zone_src: $src})"`
	}
	variables := map[string]interface{}{
		"src":   graphql.String(stats.Source),
		"dest":  graphql.String(stats.Destination),
		"hour":  timestamp(stats.Hour.Format(types.Format)),
		"count": graphql.Int(stats.Count),
	}

	err := c.c.Mutate(ctx, &mutation, variables)
	if err == nil && mutation.M.Rows != 1 {
		return errors.New("invalid amount of rows edited")
	}
	return err
}

func (c *Client) ibcTxIncrement(ctx context.Context, stats types.IbcStats) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"update_ibc_tx_hourly_stats(where: {hour: {_eq: $hour}, zone_dest: {_eq: $dest}, zone_src: {_eq: $src}}, _inc: {txs_cnt: $count})"`
	}

	variables := map[string]interface{}{
		"src":   graphql.String(stats.Source),
		"dest":  graphql.String(stats.Destination),
		"hour":  timestamp(stats.Hour.Format(types.Format)),
		"count": graphql.Int(stats.Count),
	}

	err := c.c.Mutate(ctx, &mutation, variables)
	if err == nil && mutation.M.Rows != 1 {
		return errors.New("invalid amount of rows edited")
	}

	return err
}

func (c *Client) ibcTxUpsert(ctx context.Context, stats types.IbcStats) error {
	exists, err := c.ibcTxExists(ctx, stats)
	if err != nil {
		return err
	}
	if !exists {
		return c.ibcTxAdd(ctx, stats)
	}
	return c.ibcTxIncrement(ctx, stats)
}

// IbcTxUpsert updates and creates fields according to IbcStats content in db
func (c *Client) IbcTxUpsert(ctx context.Context, stats ...types.IbcStats) error {
	for _, s := range stats {
		err := c.ibcTxUpsert(ctx, s)
		if err != nil {
			return err
		}
	}
	return nil
}
