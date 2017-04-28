// The goal here is to make a script which can connect to a storage node and
// pull out the current schemas as defined and spit them back to the user
// in dataman format.
//
// For now this will simply be something that knows how to interact with just postgres
// but once we do a split of interfaces in the storage node we should be able to use
// any storage node to do so
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Databases []string `long:"databases"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	meta := metadata.NewMeta()

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

	// TODO: better
	store, err := datasourceInstanceConfig.GetStore(func() *metadata.Meta { return nil })
	storeSchema := store.(storagenode.StorageSchemaInterface)

	for _, databasename := range opts.Databases {
		meta.Databases[databasename] = storeSchema.GetDatabase(databasename)
	}

	// TODO: sort? it'd be nice to have the files not change if there was no schema change
	bytes, _ := json.MarshalIndent(meta, "", "  ")
	fmt.Println(string(bytes))

}
