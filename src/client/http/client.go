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

func NewHTTPTransport(destination string) (*HTTPTransport, error) {
	return &HTTPTransport{
		destination: destination,
		client:      &http.Client{},
	}, nil
}

type HTTPTransport struct {
	destination string
	client      *http.Client
}

func (d *HTTPTransport) DoQuery(ctx context.Context, q *query.Query) (*query.Result, error) {
	encArgs, err := json.Marshal(q.Args)
	if err != nil {
		return nil, fmt.Errorf("Json error: %v", err)
	}
	bodyReader := bytes.NewReader(encArgs)

	// send task to node
	req, err := http.NewRequest("POST", d.destination+"/"+string(q.Type), bodyReader)
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
