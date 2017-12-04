package datamanclient

import (
	"context"

	"github.com/jacksontj/dataman/src/query"
)

// TODO: support per-query config?
// TODO support switching config in-flight? If so then we'll need to store a
// pointer to it in the context -- which would require implementing one ourself
type Client struct {
	Transport DatamanClientTransport
	// TODO: config (timeout, etc).
}

// TODO: add these convenience methods
/*
   Get(query.QueryArgs) *query.Result
   Set(query.QueryArgs) *query.Result
   Insert(query.QueryArgs) *query.Result
   Update(query.QueryArgs) *query.Result
   Delete(query.QueryArgs) *query.Result
*/

// DoQuery will execute a given query. This will return a (result, error) -- where the
// error is any transport level error (NOTE: any response errors due to the query will *not*
// be reported in this error, they will be in the normal Result.Error location)
func (d *Client) DoQuery(ctx context.Context, q *query.Query) (*query.Result, error) {
	c, cancel := context.WithCancel(ctx)
	defer cancel() // Cancel ctx as soon as handleSearch returns.

	results, err := d.Transport.DoQuery(c, q)
	if err != nil {
		return nil, err
	} else {
		return results, err
	}
}
