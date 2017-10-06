package datamanhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	queryMap := map[query.QueryType]interface{}{q.Type: q.Args}

	encQueries, err := json.Marshal(queryMap)
	if err != nil {
		return nil, fmt.Errorf("Json error: %v", err)
	}
	bodyReader := bytes.NewReader(encQueries)

	// send task to node
	req, err := http.NewRequest("POST", d.destination, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}

	// Pass the context on
	req = req.WithContext(ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error from http request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing http response: %v", err)
	}

	var result *query.Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(string(body))
		return nil, fmt.Errorf("Error unmarshaling json response: %v", err)
	} else {
		return result, nil
	}
}
