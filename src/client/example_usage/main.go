package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/client"
	"github.com/jacksontj/dataman/src/client/direct"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func doExamples(client *datamanclient.Client) {
	ret, err := client.DoQuery(
		map[query.QueryType]query.QueryArgs{
			query.Filter: map[string]interface{}{
				"db":             "example_forum",
				"collection":     "user",
				"shard_instance": "dbshard_example_forum_2",
				"filter":         map[string]interface{}{},
			},
		},
	)

	fmt.Println(ret, err)
}

func main() {
	config := &storagenode.Config{}
	configBytes, err := ioutil.ReadFile("../../storage_node/storagenode/config.yaml")
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		logrus.Fatalf("Error unmarshaling config: %v", err)
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

	var datasourceInstanceConfig *storagenode.DatasourceInstanceConfig
	for _, v := range config.Datasources {
		datasourceInstanceConfig = v
		break
	}

	// TODO: remove
	datasourceInstanceConfig.SkipProvisionTrim = true

	transport, err := datamandirect.NewStaticDatasourceInstanceTransport(datasourceInstanceConfig, meta)
	if err != nil {
		logrus.Fatalf("Error NewStaticDatasourceInstanceClient: %v", err)
	}

	client := &datamanclient.Client{Transport: transport}
	doExamples(client)

}
