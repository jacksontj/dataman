package storagenode

// A collection of tests to test the storagenode
// This includes:
//		- schema validation

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
)

// TODO: have a list of them? We want to test all of them (or become a library of tests
// that the modules can just run
func getNode() (*StorageNode, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}
	node, err := NewStorageNode(store)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// Test Functions for covering a document DB
func TestNodeDocumentDatabase(t *testing.T) {
	node, err := getNode()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}

	meta := node.Store.GetMeta()

	// TODO: move into getStore()
	// Clear the DB -- since we are going to use it
	for _, db := range meta.Databases {
		if err := node.Store.RemoveDatabase(db.Name); err != nil {
			t.Fatalf("Unable to remove DB: %v", err)
		}
	}

	// document schema tests
	databaseAdd := &metadata.Database{
		Name: "docdb",
		Tables: map[string]*metadata.Table{
			"person": &metadata.Table{
				Name: "person",
				Columns: []*metadata.TableColumn{
					&metadata.TableColumn{
						Name: "data",
						Type: metadata.Document,
						Schema: &metadata.Schema{
							Name:    "person",
							Version: 1,
							Schema: map[string]interface{}{
								"title": "Person",
								"type":  "object",
								"properties": map[string]interface{}{
									"firstName": map[string]interface{}{
										"type": "string",
									},
								},
								"required": []string{"firstName"},
							},
						},
					},
				},
			},
		},
	}

	// Add the database
	if err := node.Store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	var q map[query.QueryType]query.QueryArgs

	// Write up the query as it would look on the wire, then we can just use it
	// instead of doing all the golang object creation (since its tedious)
	queryBytes := []byte(`
	{
		"set": {
			"db": "docdb",
			"table": "person",
            "columns": {
			    "data": {
				    "lastName": "mctester"
			    }
		    }
		}
	}
	`)
	json.Unmarshal(queryBytes, &q)
	result := node.HandleQuery(q)
	if result.Error == "" {
		t.Fatalf("No error when adding an invalid document!")
	}

	queryBytes = []byte(`
	{
        "set": {
            "db": "docdb",
            "table": "person",
            "columns": {
                "data": {
                        "firstName": "tester"
                }
            }
        }
    }
	`)
	q = make(map[query.QueryType]query.QueryArgs)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", err)
	}

	queryBytes = []byte(`
	{
        "set": {
            "db": "docdb",
            "table": "person",
            "columns": {
                "data": {
                        "firstName": "otherguy"
                }
            }
        }
    }
	`)
	q = make(map[query.QueryType]query.QueryArgs)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", err)
	}

	queryBytes = []byte(`
    {
        "set": {
            "db": "docdb",
            "table": "person",
            "columns": {
                "data": {
                    "firstName": "tester",
	                "lastName": "foobar"
                }
            }
        }
    }
    `)
	q = make(map[query.QueryType]query.QueryArgs)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}

	queryBytes = []byte(`
    {
        "filter": {
            "db": "docdb",
            "table": "person",
            "columns": {
                "data": {
                    "firstName": "tester"
                }
            }
        }
    }
    `)
	q = make(map[query.QueryType]query.QueryArgs)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error != "" {
		t.Fatalf("Error when doing a valid filter(): %v", result.Error)
	}
	if len(result.Return) != 2 {
		t.Fatalf("Returns not what we expect, expected 2 got %d: %v", len(result.Return), result.Return)
	}

	// TODO: we need to get back the IDs of the documents to call delete-- otherwise it is a filter delete
	// Delete
	queryBytes = []byte(fmt.Sprintf(`
    {
        "delete": {
            "db": "docdb",
            "table": "person",
            "columns": {
                "_id": %v
            }
        }
    }
    `, result.Return[0]["_id"]))
	q = make(map[query.QueryType]query.QueryArgs)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error != "" {
		t.Fatalf("Error when doing a valid delete(): %v", result.Error)
	}

	// TODO: some other way..  we just want to make sure its 1
	queryBytes = []byte(`
    {
        "filter": {
            "db": "docdb",
            "table": "person",
            "columns": {
                "data": {
                    "firstName": "tester"
                }
            }
        }
    }
    `)
	q = make(map[query.QueryType]query.QueryArgs)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error != "" {
		t.Fatalf("Error when doing a valid filter(): %v", result.Error)
	}
	if len(result.Return) != 1 {
		t.Fatalf("Returns not what we expect, expected 1 got %d: %v", len(result.Return), result.Return)
	}

}
