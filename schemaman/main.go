// The goal here is to create a CLI which can connect to one DB and copy the schema (and data) to another one

package main

import (
	"io/ioutil"

	"github.com/jacksontj/dataman/storagenode"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var opts struct {
	ConfigFile string `long:"config" description:"path to the config file" required:"true"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	config := &Config{}
	configBytes, err := ioutil.ReadFile(opts.ConfigFile)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		logrus.Fatalf("Error unmarshaling config: %v", err)
	}
	logrus.Infof("config: %v", config)

	// Create all the datasource_instances
	datasourceInstances := make(map[string]*storagenode.DatasourceInstance)
	for name, datasourceInstanceConfig := range config.Datasources {
		if datasourceInstance, err := storagenode.NewDatasourceInstance(datasourceInstanceConfig); err == nil {
			datasourceInstances[name] = datasourceInstance
		} else {
			logrus.Fatalf("Unable to create datasource_instance %s: %v", name, err)
		}
	}

	// Do all the actions
	for i, action := range config.Actions {
		if err := action.Execute(datasourceInstances); err == nil {
			logrus.Infof("Action %d.%s completed", i, action.Action)
		} else {
			logrus.Infof("Action %d.%s Error: %v", i, action.Action, err)
		}
	}
}
