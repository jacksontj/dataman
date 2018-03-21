package datamanhttp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/stream/httpjson"
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

func (d *HTTPTransport) DoStreamQuery(ctx context.Context, q *query.Query) (*query.ResultStream, error) {
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

	reader := bufio.NewReader(resp.Body)
	buf, err := reader.ReadBytes('\n')
	if err != nil {
		resp.Body.Close() // TODO: better? we don't want defer close, as the background reader is responsible
		return nil, err
	}

	/// unmarshal the base struct
	var result *query.ResultStream
	if err := json.Unmarshal(buf, &result); err != nil {
		resp.Body.Close() // TODO: better? we don't want defer close, as the background reader is responsible
		return nil, err
	}

	// We have to use this BufCloser as the bufio.Reader pulls things into the buffer,
	// so they won't be in the resp.Body
	result.Stream = httpjson.NewClientStream(&BufCloser{reader, resp.Body})

	return result, nil
}

// TODO: move elsewhere?
type BufCloser struct {
	*bufio.Reader
	c io.ReadCloser
}

func (b *BufCloser) Close() error { return b.c.Close() }
