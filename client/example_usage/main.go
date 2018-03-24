package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/jacksontj/dataman/client"
	"github.com/jacksontj/dataman/client/direct"
	"github.com/jacksontj/dataman/client/http"
	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode"
	"github.com/jacksontj/dataman/storagenode/datasource"
	"github.com/jacksontj/dataman/storagenode/metadata"
	"github.com/sirupsen/logrus"
)

func doExamples(client *datamanclient.Client) error {
	q := &query.Query{
		query.Filter,
		map[string]interface{}{
			"db":             "example_forum",
			"collection":     "user",
			"shard_instance": "dbshard_example_forum_7_1",
			"filter":         map[string]interface{}{},
		},
	}

	ret, err := client.DoQuery(context.Background(), q)

	if err != nil {
		fmt.Println(ret, err)
	}
	return err
}

// Example of using direct with a static config file
func directStatic() {
	config, err := storagenode.DatasourceInstanceConfigFromFile("datasourceinstance.yaml")
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}
	logrus.Infof("config: %v", config)

	// Load meta
	meta := &metadata.Meta{}
	metaBytes, err := ioutil.ReadFile("schema.json")
	if err != nil {
		logrus.Fatalf("Error loading schema: %v", err)
	}
	err = json.Unmarshal([]byte(metaBytes), meta)
	if err != nil {
		logrus.Fatalf("Error loading meta: %v", err)
	}

	// TODO: remove
	config.SkipProvisionTrim = true

	transport, err := datamandirect.NewStaticDatasourceInstanceTransport(config, meta)
	if err != nil {
		logrus.Fatalf("Error NewStaticDatasourceInstanceClient: %v", err)
	}

	client := &datamanclient.Client{Transport: transport}
	if err := doExamples(client); err != nil {
		fmt.Println("error with directStatic")
	} else {
		fmt.Println("directStatic success")
	}

}

// Example of using direct with dynamically finding the schema on startup
func directDynamic() {
	config, err := storagenode.DatasourceInstanceConfigFromFile("datasourceinstance.yaml")
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}
	logrus.Infof("config: %v", config)

	// TODO: remove
	config.SkipProvisionTrim = true

	meta := metadata.NewMeta()
	// Note: since we are soely doing schema *export* we don't define a meta func
	// this means that all writes will fail as there is no schema to compare to
	store, err := config.GetStore(nil)
	storeSchema := store.(datasource.SchemaInterface)

	for _, database := range storeSchema.ListDatabase(context.Background()) {
		meta.Databases[database.Name] = database
	}

	transport, err := datamandirect.NewStaticDatasourceInstanceTransport(config, meta)
	if err != nil {
		logrus.Fatalf("Error NewStaticDatasourceInstanceClient: %v", err)
	}

	client := &datamanclient.Client{Transport: transport}
	if err := doExamples(client); err != nil {
		fmt.Println("error with directDynamic")
	} else {
		fmt.Println("directDynamic success")
	}

}

func http() {
	transport, err := datamanhttp.NewHTTPDatamanClient("http://127.0.0.1:8080/v1/data/raw")
	if err != nil {
		logrus.Fatalf("Error NewHTTPDatamanClient: %v", err)
	}

	client := &datamanclient.Client{Transport: transport}
	if err := doExamples(client); err != nil {
		fmt.Println("error with http")
	} else {
		fmt.Println("http success")
	}

}

func main() {
	directStatic()
	directDynamic()
	http()
}
