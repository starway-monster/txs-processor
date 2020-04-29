package graphql

import (
	"context"

	"github.com/mapofzones/txs-processor/types"
	"github.com/shurcooL/graphql"
)

// AddBlock tells database that the block is processed
func (p *Client) AddBlock(ctx context.Context, b types.Block) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"insert_blocks_log(objects: {height: $height, zone: $zone})"`
	}

	variables := map[string]interface{}{
		"zone":   graphql.String(b.ChainID),
		"height": graphql.Int(b.Height),
	}

	return p.c.Mutate(ctx, &mutation, variables)
}
