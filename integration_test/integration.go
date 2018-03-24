package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"net/http"
	_ "net/http/pprof"

	"github.com/jacksontj/dataman/client"
	"github.com/jacksontj/dataman/client/direct"
	"github.com/jacksontj/dataman/client/http"
	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/routernode"
	"github.com/jacksontj/dataman/routernode/metadata"
	"github.com/jacksontj/dataman/storagenode"
	"github.com/jacksontj/dataman/tasknode"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var datamanClientTransport string

func init() {
	flag.StringVar(&datamanClientTransport, "dataman-client-transport", "direct", "Which transport to use for the dataman client")
	flag.Parse()
}

type Data map[string]map[string][]map[string]interface{}

// For this we assume these are empty and we can do whatever we want to them!
func RunIntegrationTests(t *testing.T, task *tasknode.TaskNode, router *routernode.RouterNode, datasource *storagenode.DatasourceInstance) {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Find all tests
	files, err := ioutil.ReadDir("tests")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		// TODO: subtest stuff
		t.Run(file.Name(), func(t *testing.T) {
			runIntegrationTest(file.Name(), t, task, router, datasource)
		})
	}
}

func runIntegrationTest(testDir string, t *testing.T, task *tasknode.TaskNode, router *routernode.RouterNode, datasource *storagenode.DatasourceInstance) {
	// Get the transport requested from the CLI
	getTransport := func() datamanclient.DatamanClientTransport {
		switch datamanClientTransport {
		case "direct":
			return datamandirect.NewRouterTransport(router)

		case "http":
			transport, _ := datamanhttp.NewHTTPTransport("http://127.0.0.1" + router.Config.HTTP.Addr + "/v1/data/raw")
			return transport

		default:
			log.Fatalf("Unknown datman-client-transport: %s", datamanClientTransport)
			return nil
		}
	}

	// TODO: use client for schema manipulation too
	// Load the various files
	schema := make(map[string]*metadata.Database)
	schemaBytes, err := ioutil.ReadFile(path.Join("tests", testDir, "/schema.json"))
	if err != nil {
		t.Fatalf("Unable to read schema for test %s: %v", testDir, err)
	}
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
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

	client := datamanclient.Client{getTransport()}

	// Load data
	data := make(Data)
	dataString, err := ioutil.ReadFile(path.Join("tests", testDir, "/data.json"))
	if err != nil {
		t.Fatalf("Unable to read data for test %s: %v", testDir, err)
	}
	if err := json.Unmarshal([]byte(dataString), &data); err != nil {
		t.Fatalf("Unable to load data for test %s: %v", testDir, err)
	}

	for databaseName, collectionMap := range data {
		for collectionName, recordList := range collectionMap {
			for _, record := range recordList {
				// TODO: switch to using the client interface (with CLI flag for which one to use)
				result, err := client.DoQuery(context.Background(), &query.Query{
					Type: query.Insert,
					Args: query.QueryArgs{
						DB:         databaseName,
						Collection: collectionName,
						Record:     record,
					},
				})
				if err != nil {
					t.Fatalf("Transport error on client: %v", err)
				}
				if result.ValidationError != nil {
					t.Fatalf("Valdiation error loading data into %s.%s: %v", databaseName, collectionName, result.ValidationError)
				}
				if err := result.Err(); err != nil {
					t.Fatalf("Error loading data into %s.%s: %v", databaseName, collectionName, err)
				}
			}
		}
	}
	testsDir := path.Join("tests", testDir, "/")

	walkFunc := func(fpath string, info os.FileInfo, err error) error {
		// If its a directory, skip it-- we'll let something else grab it
		if !info.IsDir() {
			return nil
		}

		// if this is a test directory it must have a query.json file
		// Load the query
		q := &query.Query{}
		queryBytes, err := ioutil.ReadFile(path.Join(fpath, "query.json"))
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			t.Fatalf("Unable to read queryBytes for test %s.%s: %v", testDir, info.Name(), err)
		}

		relFilePath, err := filepath.Rel(testsDir, fpath)
		if err != nil {
			t.Fatalf("Error getting relative path? Shouldn't be possible: %v", err)
		}
		t.Run(relFilePath, func(t *testing.T) {

			if err := json.Unmarshal([]byte(queryBytes), &q); err != nil {
				t.Fatalf("Unable to load queries for test %s.%s: %v", testDir, info.Name(), err)
			}

			var queryResult interface{}

			// Run the type of query it is
			switch q.Type {
			case query.FilterStream:
				resultStream, err := client.DoStreamQuery(context.Background(), q)
				if err != nil {
					t.Fatalf("Transport error on client: %v", err)
				}

				streamResult := []interface{}{resultStream}

				// TODO: wrap the results in a struct for marshaling?
				for {
					if val, err := resultStream.Recv(); err != nil {
						streamResult = append(streamResult, err.Error())
						break
					} else {
						streamResult = append(streamResult, val)
					}
				}
				queryResult = streamResult

			default:
				// Run the query
				result, err := client.DoQuery(context.Background(), q)
				if err != nil {
					t.Fatalf("Transport error on client: %v", err)
				}
				queryResult = result
			}

			// Check result

			// write out results
			resultPath := path.Join(fpath, "result.json")
			resultBytes, _ := json.MarshalIndent(queryResult, "", "  ")
			ioutil.WriteFile(resultPath, resultBytes, 0644)

			// compare against baseline if it exists
			baselinePath := path.Join(fpath, "baseline.json")
			baselineResultBytes, err := ioutil.ReadFile(baselinePath)
			if err != nil {
				t.Skip("No baseline.json found, skipping comparison")
			} else {
				baselineResultBytes = bytes.TrimSpace(baselineResultBytes)
				resultBytes = bytes.TrimSpace(resultBytes)
				if !bytes.Equal(baselineResultBytes, resultBytes) {
					dmp := diffmatchpatch.New()
					diffs := dmp.DiffMain(string(baselineResultBytes), string(resultBytes), false)
					t.Fatalf("Mismatch of results and baseline!\n%s", dmp.DiffPrettyText(diffs))
				}
			}
		})
		return nil
	}

	if err := filepath.Walk(testsDir, walkFunc); err != nil {
		t.Errorf("Error walking: %v", err)
	}

	// Assuming we finished properly, lets remove the things we added
	for _, database := range schema {
		if err := task.EnsureDoesntExistDatabase(context.Background(), database.Name); err != nil {
			t.Fatalf("Unable to remove in test %s for database %s: %v", testDir, database.Name, err)
		}
	}

	// Block the router waiting on an update from the tasknode
	router.FetchMeta()

}
