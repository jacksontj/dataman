// This is an attempt at making test cases that are just config file defined
// since the majority of the use-cases will be something like "define schema, run query"
package storagenode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
)

type StoreTestFixture struct {
	// TODO: map of dbname -> db
	Schema *metadata.Database `json:"schema"`

	// Map of db -> table -> item
	SeedData map[string]map[string]map[string]interface{} `json:"seed_data"`

	Queries []StoreTestQuery `json:"queries"`
}

type StoreTestResult struct {
	NumResults int                      `json:"num_results"`
	Returns    []map[string]interface{} `json:"returns"`
	HasError   bool                     `json:"error"`
}

func (s *StoreTestResult) Compare(r *query.Result) error {
	if s.NumResults > 0 {
		if len(r.Return) != s.NumResults {
			return fmt.Errorf("result has wrong number of returns expected=%d actual=%d", s.NumResults, len(r.Return))
		}
	}

	if s.HasError != (r.Error != "") {
		return fmt.Errorf("Mismatched error expected=%v actual=%v", s.HasError, r.Error != "")
	}

	// compare the returns as subsets
	if s.Returns != nil {
		if len(s.Returns) != len(r.Return) {
			return fmt.Errorf("Mismatched return lens expected=%d actual=%d", len(s.Returns), len(r.Return))
		}
		for i, ret := range s.Returns {
			for k, v := range ret {
				actualValue, ok := r.Return[i][k]
				if !ok {
					fmt.Errorf("Missing expected key %v", k)
				}
				if !reflect.DeepEqual(v, actualValue) {
					fmt.Errorf("Mismatched values expected=%v actual=%v", v, actualValue)
				}
			}
		}
	}
	return nil
}

type StoreTestQuery struct {
	// A query to run
	Query map[query.QueryType]query.QueryArgs `json:"query"`

	// A subset of the actual result-- pinning what we care about
	Result StoreTestResult `json:"result"`
}

// TODO: test indexes
// Test Functions for covering a document DB
func TestDBSimple(t *testing.T) {
	node, err := getNode()
	if err != nil {
		t.Fatalf("Unable to create test store: %v", err)
	}
	resetStore(node.Store)

	var test StoreTestFixture
	bytes, err := ioutil.ReadFile("test.json")
	if err != nil {
		t.Fatalf("Err reading json: %v", err)
	}
	err = json.Unmarshal(bytes, &test)
	if err != nil {
		t.Fatalf("Err loading json: %v", err)
	}

	err = node.Store.AddDatabase(test.Schema)
	if err != nil {
		t.Fatalf("Unable to add database: %v", err)
	}

	// Now we run all the queries
	for _, q := range test.Queries {
		result := node.HandleQuery(q.Query)
		if err := q.Result.Compare(result); err != nil {
			t.Fatalf("Error when running %v: %v", q.Query, err)
		}
	}

}
