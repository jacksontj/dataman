package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/client/direct"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	// Load config
	config := &storagenode.Config{}
	configBytes, err := ioutil.ReadFile("../../../storage_node/storagenode/config.yaml")
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
	err = json.Unmarshal([]byte(storagenode.GetSchema()), meta)
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

	client, err := datamandirect.NewStaticDatasourceInstanceClient(datasourceInstanceConfig, meta)
	if err != nil {
		logrus.Fatalf("Error NewStaticDatasourceInstanceClient: %v", err)
	}

	ret := client.DoQuery(
		map[query.QueryType]query.QueryArgs{
			query.Get: map[string]interface{}{
				"db":             "dataman_storage",
				"shard_instance": "public",
				"collection":     "field_type",
				"_id":            1,
			},
		},
	)

	fmt.Println(ret)

}
