// The goal here is to make a script which can connect to a storage node and
// pull out the current schemas as defined and spit them back to the user
// in dataman format.
//
// For now this will simply be something that knows how to interact with just postgres
// but once we do a split of interfaces in the storage node we should be able to use
// any storage node to do so
package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var opts struct {
	Schema    string   `long:"schema"`
	Databases []string `long:"databases"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	// Load schema file
	schemaBytes, err := ioutil.ReadFile(opts.Schema)
	if err != nil {
		logrus.Fatalf("unable to find config: %v", err)
	}
	meta := metadata.Meta{}
	err = json.Unmarshal([]byte(schemaBytes), &meta)
	if err != nil {
		logrus.Fatalf("invalid schema: %v", err)
	}

	// Create datasourceInstance (mutablemetastore)
	// TODO: actually have these come through CLI args or something
	config := &storagenode.Config{}
	configBytes, err := ioutil.ReadFile("../storage_node/storagenode/config.yaml")
	if err != nil {
		logrus.Fatalf("unable to find config: %v", err)
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		logrus.Fatalf("invalid config: %v", err)
	}

	var datasourceInstanceConfig *storagenode.DatasourceInstanceConfig
	for _, c := range config.Datasources {
		datasourceInstanceConfig = c
		break
	}

	// populate mutablemetastore with schema
	dsi, err := storagenode.NewDatasourceInstanceDefault(datasourceInstanceConfig)
	if err != nil {
		logrus.Fatalf("unable to initialze datasourceinstancedefault: %v", err)
	}

	for _, db := range meta.Databases {
		if err := dsi.MutableMetaStore.EnsureExistsDatabase(context.Background(), db); err != nil {
			logrus.Fatalf("Unable to populate mutablemetastore with %s: %v", db.Name, err)
		}
	}
}
