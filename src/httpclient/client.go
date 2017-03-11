package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
)

// Take a query and send it to a given destination
func Query(ip string, port int, queries []*query.Query) ([]*query.Result, error) {
	url := fmt.Sprintf("http://%s:%d/v1/data/raw", ip, port)

	// TODO: pass in? Or options?
	client := &http.Client{}

	// TODO: better marshalling
	queriesMap := make([]map[query.QueryType]interface{}, len(queries))
	for i, q := range queries {
		queriesMap[i] = map[query.QueryType]interface{}{q.Type: q.Args}
	}

	encQueries, err := json.Marshal(queriesMap)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(encQueries)

	// send task to node
	req, err := http.NewRequest(
		"POST",
		url,
		bodyReader,
	)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	results := make([]*query.Result, len(queries))
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func QuerySingle(ip string, port int, q *query.Query) (*query.Result, error) {
	if results, err := Query(ip, port, []*query.Query{q}); err == nil {
		return results[0], nil
	} else {
		return nil, err
	}
}
