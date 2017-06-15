package datamanhttp

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
)

func errorSlice(count int, err string) []*query.Result {
	errors := make([]*query.Result, count)
	for i := 0; i < count; i++ {
		errors[i] = &query.Result{Error: err}
	}
	return errors
}

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

func (d *HTTPDatamanClient) DoQuery(q map[query.QueryType]query.QueryArgs) *query.Result {
	return d.DoQueries([]map[query.QueryType]query.QueryArgs{q})[0]
}

func (d *HTTPDatamanClient) DoQueries(queries []map[query.QueryType]query.QueryArgs) []*query.Result {
	// TODO: better marshalling
	queriesMap := make([]map[query.QueryType]interface{}, len(queries))
	for i, q := range queries {
		for k, v := range q {
			queriesMap[i] = map[query.QueryType]interface{}{k: v}
		}
	}

	encQueries, err := json.Marshal(queriesMap)
	if err != nil {
		return errorSlice(len(queries), err.Error())
	}
	bodyReader := bytes.NewReader(encQueries)

	// send task to node
	req, err := http.NewRequest(
		"POST",
		d.destination+"data/raw",
		bodyReader,
	)
	if err != nil {
		return errorSlice(len(queries), err.Error())
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return errorSlice(len(queries), err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errorSlice(len(queries), err.Error())
	}

	results := make([]*query.Result, len(queries))
	err = json.Unmarshal(body, &results)
	if err != nil {
		return errorSlice(len(queries), err.Error())
	}

	return results
}
