package routernode

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node/metadata"
)

/*
// Get a result from at least one replica per shard
func MultiQuery(shards []*metadata.DatastoreShard, queries []*query.Query) ([]*query.Result, error) {
	mergedResults := make([]*query.Result, len(queries))

	for _, shard := range shards {
		// TODO: on error- we can try another replica (since each replica should have the
		// same exact data
		results, err := Query(shard.GetReplica(), queries)
		if err != nil {
			return nil, err
		}

		for i, result := range results {
			if mergedResults[i] == nil {
				mergedResults[i] = result
			} else {
				// Merge return lists
				mergedResults[i].Return = append(mergedResults[i].Return, result.Return...)
				// TODO: handle error and meta merging as well
			}
		}
	}

	return mergedResults, nil
}

// Get a result from at least one replica per shard
func MultiQuerySingle(shards []*metadata.DatastoreShard, q *query.Query) (*query.Result, error) {
	if ret, err := MultiQuery(shards, []*query.Query{q}); err == nil {
		return ret[0], nil
	} else {
		return nil, err
	}
}

*/

// Take a query and send it to a given destination
func Query(datasource *metadata.DatasourceInstance, queries []*query.Query) ([]*query.Result, error) {
	// TODO: pass in? Or options?
	// TODO: these should be associated with the storage_node (since that is what we are talking through)
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
		datasource.GetURL(),
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

func QuerySingle(datasource *metadata.DatasourceInstance, q *query.Query) (*query.Result, error) {
	if results, err := Query(datasource, []*query.Query{q}); err == nil {
		return results[0], nil
	} else {
		return nil, err
	}
}
