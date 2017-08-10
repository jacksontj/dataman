package integrationtest

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jacksontj/dataman/src/router_node"
	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/jacksontj/dataman/src/task_node"
	yaml "gopkg.in/yaml.v2"
)

func Setup() (*tasknode.TaskNode, *routernode.RouterNode, *storagenode.DatasourceInstance) {
	// TaskNode (with static config?)
	tasknodeConfig := &tasknode.Config{}
	tasknodeConfigBytes, err := ioutil.ReadFile("../task_node/tasknode/config.yaml")
	if err != nil {
		logrus.Fatalf("Error loading tasknodeConfig: %v", err)
	}
	err = yaml.Unmarshal([]byte(tasknodeConfigBytes), &tasknodeConfig)
	if err != nil {
		logrus.Fatalf("Error unmarshaling tasknodeConfig: %v", err)
	}

	taskNode, err := tasknode.NewTaskNode(tasknodeConfig)
	if err != nil {
		logrus.Fatalf("Error creating tasknode: %v", err)
	}

	go taskNode.Start()
	time.Sleep(time.Second) // TODO: remove-- need a mechanism to know its ready

	// Router (with static config?)
	routernodeConfig := &routernode.Config{}
	routernodeConfigBytes, err := ioutil.ReadFile("../router_node/routernode/config.yaml")
	if err != nil {
		logrus.Fatalf("Error loading routernodeConfig: %v", err)
	}
	err = yaml.Unmarshal([]byte(routernodeConfigBytes), &routernodeConfig)
	if err != nil {
		logrus.Fatalf("Error unmarshaling routernodeConfig: %v", err)
	}

	routerNode, err := routernode.NewRouterNode(routernodeConfig)
	if err != nil {
		logrus.Fatalf("Error creating routernode: %v", err)
	}
	go routerNode.Start()

	// Datasource
	storagenodeConfig := &storagenode.Config{}
	storagenodeConfigBytes, err := ioutil.ReadFile("../storage_node/storagenode/config.yaml")
	if err != nil {
		logrus.Fatalf("Error loading storagenodeConfig: %v", err)
	}
	err = yaml.Unmarshal([]byte(storagenodeConfigBytes), &storagenodeConfig)
	if err != nil {
		logrus.Fatalf("Error unmarshaling storagenodeConfig: %v", err)
	}

	for _, datasourceConfig := range storagenodeConfig.Datasources {
		datasourceConfig.SkipProvisionTrim = true
	}

	storageNode, err := storagenode.NewStorageNode(storagenodeConfig)
	if err != nil {
		logrus.Fatalf("Unable to create storagenode: %v", err)
	}
	go storageNode.Start()

	var datasourceInstance *storagenode.DatasourceInstance
	for _, dsi := range storageNode.Datasources {
		datasourceInstance = dsi
		break
	}

	// Clear things
	// Clear router
	for _, database := range taskNode.GetMeta().Databases {
		if err := taskNode.EnsureDoesntExistDatabase(context.Background(), database.Name); err != nil {
			logrus.Fatalf("Unable to clear database from tasknode: %v", err)
		}
	}

	// Clear datasourceInstance
	for _, database := range datasourceInstance.GetMeta().Databases {
		if err := taskNode.EnsureDoesntExistDatabase(context.Background(), database.Name); err != nil {
			logrus.Fatalf("Unable to clear database from tasknode: %v", err)
		}
	}

	return taskNode, routerNode, datasourceInstance
}
