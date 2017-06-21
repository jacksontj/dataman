package datamanclient

import (
	"context"
	"time"

	"github.com/jacksontj/dataman/src/query"
)

// TODO: support per-query config?
// TODO support switching config in-flight? If so then we'll need to store a
// pointer to it in the context -- which would require implementing one ourself
type Client struct {
	Transport DatamanClientTransport
	// TODO: config
}

// TODO: add these convenience methods
/*
   Get(query.QueryArgs) *query.Result
   Set(query.QueryArgs) *query.Result
   Insert(query.QueryArgs) *query.Result
   Update(query.QueryArgs) *query.Result
   Delete(query.QueryArgs) *query.Result
*/

func (d *Client) DoQuery(q map[query.QueryType]query.QueryArgs) (*query.Result, error) {
	timeout := time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // Cancel ctx as soon as handleSearch returns.

	results, err := d.Transport.DoQueries(ctx, []map[query.QueryType]query.QueryArgs{q})
	if err != nil {
		return nil, err
	} else {
		return results[0], err
	}
}

func (d *Client) DoQueries(q []map[query.QueryType]query.QueryArgs) ([]*query.Result, error) {
	timeout := time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // Cancel ctx as soon as handleSearch returns.

	return d.Transport.DoQueries(ctx, q)
}
