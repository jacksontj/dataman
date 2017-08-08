package integrationtest

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/router_node"
	"github.com/jacksontj/dataman/src/router_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/task_node"
)

type Data map[string]map[string][]map[string]interface{}

// For this we assume these are empty and we can do whatever we want to them!
func RunIntegrationTests(t *testing.T, task *tasknode.TaskNode, router *routernode.RouterNode, datasource *storagenode.DatasourceInstance) {

	// Find all tests
	files, err := ioutil.ReadDir("tests")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		// TODO: subtest stuff
		runIntegrationTest(file.Name(), t, task, router, datasource)
	}
}

func runIntegrationTest(testDir string, t *testing.T, task *tasknode.TaskNode, router *routernode.RouterNode, datasource *storagenode.DatasourceInstance) {
	// Load the various files
	schema := make(map[string]*metadata.Database)
	schemaString, err := ioutil.ReadFile("tests/" + testDir + "/schema.json")
	if err != nil {
		t.Fatalf("Unable to read schema for test %s: %v", testDir, err)
	}
	if err := json.Unmarshal([]byte(schemaString), &schema); err != nil {
		t.Fatalf("Unable to load schema for test %s: %v", testDir, err)
	}

	// Load the schema
	for _, database := range schema {
		if err := task.EnsureExistsDatabase(context.Background(), database); err != nil {
			t.Fatalf("Unable to ensureSchema in test %s for database %s: %v", testDir, database.Name, err)
		}
	}
	
	// Block the router waiting on an update from the tasknode
	router.FetchMeta()

	// Load data
	data := make(Data)
	dataString, err := ioutil.ReadFile("tests/" + testDir + "/data.json")
	if err != nil {
		t.Fatalf("Unable to read data for test %s: %v", testDir, err)
	}
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		t.Fatalf("Unable to load data for test %s: %v", testDir, err)
	}

	for databaseName, collectionMap := range data {
		for collectionName, recordList := range collectionMap {
			for _, record := range recordList {
				result := router.HandleQuery(context.Background(), &query.Query{
					Type: query.Insert,
					Args: map[string]interface{}{
						"db":         databaseName,
						"collection": collectionName,
						"record":     record,
					},
				})
				if result.ValidationError != nil {
					t.Fatalf("Valdiation error loading data into %s.%s: %v", databaseName, collectionName, result.ValidationError)
				}
				if result.Error != "" {
					t.Fatalf("Error loading data into %s.%s: %v", databaseName, collectionName, result.Error)
				}
			}
		}
	}

}
