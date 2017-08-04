package datamanhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
)

func NewHTTPDatamanClient(destination string) (*HTTPDatamanClient, error) {
	return &HTTPDatamanClient{
		destination: destination,
		client:      &http.Client{},
	}, nil
}

type HTTPDatamanClient struct {
	destination string
	client      *http.Client
}

func (d *HTTPDatamanClient) DoQuery(ctx context.Context, q *query.Query) (*query.Result, error) {
	// TODO: better marshalling
	queryArgs := make(map[string]interface{})
	for k, v := range q.Args {
		queryArgs[k] = v
	}
	queryMap := map[query.QueryType]interface{}{q.Type: queryArgs}

	encQueries, err := json.Marshal(queryMap)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(encQueries)

	// send task to node
	req, err := http.NewRequest(
		"POST",
		d.destination+"data/raw",
		bodyReader,
	)
	if err != nil {
		return nil, err
	}

	// Pass the context on
	req.WithContext(ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result *query.Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
