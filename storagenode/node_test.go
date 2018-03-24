package storagenode

// A collection of tests to test the storagenode
// This includes:
//		- schema validation
/*
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

// TODO: have a list of them? We want to test all of them (or become a library of tests
// that the modules can just run
func getNode() (*StorageNode, error) {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("storagenode/config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		return nil, err
	}
	node, err := NewStorageNode(config)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func resetNode(node *StorageNode) error {
	meta := node.GetMeta()

	for _, db := range meta.Databases {
		if err := node.RemoveDatabase(db.Name); err != nil {
			return err
		}
	}

	return nil
}

// Test Functions for covering a document DB
func TestNodeDocumentDatabase(t *testing.T) {
	node, err := getNode()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}

	resetNode(node)
	defer func() { resetNode(node) }()

	// document schema tests
	databaseAdd := &metadata.Database{
		Name: "docdb",
		Collections: map[string]*metadata.Collection{
			"person": &metadata.Collection{
				Name: "person",
				Fields: []*metadata.Field{
					&metadata.Field{
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
	if err := node.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	var q map[query.QueryType]query.QueryArgs

	// Write up the query as it would look on the wire, then we can just use it
	// instead of doing all the golang object creation (since its tedious)
	queryBytes := []byte(`
	{
		"insert": {
			"db": "docdb",
			"collection": "person",
            "record": {
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

	// Valid document, but including a metadata field (which we shouldn't allow)
	queryBytes = []byte(`
	{
		"insert": {
			"db": "docdb",
			"collection": "person",
            "record": {
                "_id": 12345,
			    "data": {
				    "firstName": "tester"
			    }
		    }
		}
	}
	`)
	json.Unmarshal(queryBytes, &q)
	result = node.HandleQuery(q)
	if result.Error == "" {
		t.Fatalf("No error when adding a record which includes a '_' field!")
	}

	queryBytes = []byte(`
	{
        "insert": {
            "db": "docdb",
            "collection": "person",
            "record": {
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
        "insert": {
            "db": "docdb",
            "collection": "person",
            "record": {
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
        "insert": {
            "db": "docdb",
            "collection": "person",
            "record": {
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
            "collection": "person",
            "record": {
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
	if len(result.Return) != 3 {
		t.Fatalf("Returns not what we expect, expected 3 got %d: %v", len(result.Return), result.Return)
	}

	// TODO: we need to get back the IDs of the documents to call delete-- otherwise it is a filter delete
	// Delete
	queryBytes = []byte(fmt.Sprintf(`
    {
        "delete": {
            "db": "docdb",
            "collection": "person",
            "_id": %v
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
            "collection": "person",
            "record": {
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

}
*/
