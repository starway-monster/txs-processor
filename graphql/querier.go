package graphql

import "github.com/shurcooL/graphql"

// Client is a struct that can fetch relevant data through graphql
type Client struct {
	c *graphql.Client
}

// NewClient returns abstraction over graphql api given an endpoint
func NewClient(endpoint string) *Client {
	return &Client{
		c: graphql.NewClient(endpoint, nil),
	}
}
