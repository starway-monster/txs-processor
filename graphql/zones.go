package graphql

import (
	"context"
	"sync"

	"github.com/shurcooL/graphql"
)

// this is in-memory cache of zones, so weo don't have to fetch
// it from graphql on every query
var zones map[string]string
var lock sync.RWMutex

// Zones returns all the zones registered in db
func (p *Client) Zones(ctx context.Context) ([]string, error) {
	var query struct {
		Zones []struct {
			ChainID graphql.String `graphql:"chain_id"`
		}
	}

	if err := p.c.Query(ctx, &query, nil); err != nil {
		return nil, err
	}
	out := make([]string, len(query.Zones))
	for i, v := range query.Zones {
		out[i] = string(v.ChainID)
	}
	return out, nil
}

// ZoneExists returns true if given zone is present
func (p *Client) ZoneExists(ctx context.Context, chainID string) (bool, error) {
	// if it exists in chache, it exists
	lock.RLock()
	if _, ok := zones[chainID]; ok {
		lock.RUnlock()
		return true, nil
	}
	lock.RUnlock()

	var query struct {
		Zones []struct {
			ChainID graphql.String `graphql:"chain_id"`
		} `graphql:"zones(where: {chain_id: {_eq: $id}})"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(chainID),
	}

	if err := p.c.Query(context.Background(), &query, variables); err != nil {
		return false, err
	}

	return len(query.Zones) != 0, nil
}

// ZoneName returns name stored in bd which is related to the chainID
func (p *Client) ZoneName(ctx context.Context, chainID string) (string, error) {
	// if it exists in chache, fetch it from there
	lock.RLock()
	if zone, ok := zones[chainID]; ok {
		lock.RUnlock()
		return zone, nil
	}
	lock.RUnlock()

	var query struct {
		Zones []struct {
			Name graphql.String `graphql:"name"`
		} `graphql:"zones(where: {chain_id: {_eq: $id}})"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(chainID),
	}

	if err := p.c.Query(context.Background(), &query, variables); err != nil {
		return "", err
	}
	if len(query.Zones) == 0 {
		return "", nil
	}

	lock.RLock()
	if zones == nil {
		zones = make(map[string]string)
	}
	zones[chainID] = string(query.Zones[0].Name)
	lock.RUnlock()
	return string(query.Zones[0].Name), nil
}

// AddZone creates a new zone with given
func (p *Client) AddZone(ctx context.Context, chainID string, enabled bool) error {
	var mutation struct {
		M struct {
			Rows graphql.Int `graphql:"affected_rows"`
		} `graphql:"insert_zones(objects: {chain_id: $id, name: $id, is_enabled: $enabled})"`
	}

	variables := map[string]interface{}{
		"id":      graphql.String(chainID),
		"enabled": graphql.Boolean(enabled),
	}

	return p.c.Mutate(ctx, &mutation, variables)
}
