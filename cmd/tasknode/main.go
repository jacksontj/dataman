package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"

	"github.com/jacksontj/dataman/tasknode"
)

var opts struct {
	ConfigFile string `long:"config" description:"path to the config file"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	// load the config file
	config := &tasknode.Config{}
	configBytes, err := ioutil.ReadFile(opts.ConfigFile)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		logrus.Fatalf("Error unmarshaling config: %v", err)
	}
	logrus.Infof("config: %v", config)

	taskNode, err := tasknode.NewTaskNode(config)
	if err != nil {
		logrus.Fatalf("Unable to create RouterNode: %v", err)
	}

	// initialize the http api (since at this point we are ready to go!
	taskNode.Start()
}
