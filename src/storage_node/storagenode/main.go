package main

import (
	"database/sql"
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"

	"github.com/jacksontj/dataman/src/storage_node"
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
	config := &storagenode.Config{}
	configBytes, err := ioutil.ReadFile(opts.ConfigFile)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		logrus.Fatalf("Error unmarshaling config: %v", err)
	}
	logrus.Infof("config: %v", config)

	// Get the actual StorageNode
	node := config.StorageNodeType.Get()
	if node == nil {
		logrus.Fatalf("Invalid storage_type defined: %s", config.StorageNodeType)
	}

	if err := node.Init(config.StorageConfig); err != nil {
		logrus.Fatal("Error loading storage_config: %v", err)
	}

	// initialize the http api (since at this point we are ready to go!
	router := httprouter.New()
	api := storagenode.NewHTTPApi(node)
	api.Start(router)

	http.ListenAndServe(config.HTTP.Addr, router)

	logrus.Infof("%v", node)

	db, err := sql.Open("postgres", config.PGString)
	checkErr(err)
	defer db.Close()

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
