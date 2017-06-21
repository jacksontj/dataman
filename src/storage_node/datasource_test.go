package storagenode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"

	"github.com/jacksontj/dataman/src/storage_node/metadata"

	"gopkg.in/yaml.v2"
)

func getDatasourceInstance() (*DatasourceInstance, error) {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("storagenode/config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		return nil, err
	}

	var datasourceInstanceConfig *DatasourceInstanceConfig
	for _, c := range config.Datasources {
		datasourceInstanceConfig = c
		break
	}

	return NewDatasourceInstanceDefault(datasourceInstanceConfig)
}

func resetDatasourceInstance(datasourceInstance *DatasourceInstance) error {
	meta := datasourceInstance.GetMeta()

	for dbname, _ := range meta.Databases {
		if err := datasourceInstance.EnsureDoesntExistDatabase(dbname); err != nil {
			return err
		}
	}

	return nil
}

func TestDatasource_Database(t *testing.T) {
	datasourceInstance, err := getDatasourceInstance()
	if err != nil {
		t.Fatalf("Unable to get datasourceInstance: %v", err)
	}

	if err := resetDatasourceInstance(datasourceInstance); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	testMeta, err := getTestMeta()
	if err != nil {
		logrus.Fatalf("Error loading test meta: %v", err)
	}

	// Insert the meta -- here the provision state is all 0
	if err := datasourceInstance.EnsureExistsDatabase(testMeta.Databases["example_forum"]); err != nil {
		b, _ := json.Marshal(testMeta.Databases["example_forum"])
		fmt.Printf("%s\n", b)
		t.Fatalf("Error ensuring DB: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(testMeta, datasourceInstance.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, datasourceInstance.GetMeta())
	}

	// Remove it all
	if err := datasourceInstance.EnsureDoesntExistDatabase("example_forum"); err != nil {
		t.Fatalf("Error EnsureDoesntExistDatabase: %v", err)
	}

	// TODO: check
}

func TestDatasource_ShardInstance(t *testing.T) {
	datasourceInstance, err := getDatasourceInstance()
	if err != nil {
		t.Fatalf("Unable to get datasourceInstance: %v", err)
	}

	if err := resetDatasourceInstance(datasourceInstance); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	testMeta, err := getTestMeta()
	if err != nil {
		logrus.Fatalf("Error loading test meta: %v", err)
	}

	db := &metadata.Database{Name: "example_forum"}
	// Insert the db
	if err := datasourceInstance.EnsureExistsDatabase(db); err != nil {
		t.Fatalf("Error ensuring DB: %v", err)
	}

	// set the DB id -- so the compare works
	testMeta.Databases["example_forum"].ID = db.ID

	shardInstance := testMeta.Databases["example_forum"].ShardInstances["dbshard_example_forum_2"]

	// Ensure the shardInstance
	if err := datasourceInstance.EnsureExistsShardInstance(db, shardInstance); err != nil {
		t.Fatalf("Error ensuring shardInstance: %v", err)
	}

	testMeta.Databases["example_forum"].ProvisionState = metadata.Active
	// Check
	if !metaEqual(testMeta, datasourceInstance.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, datasourceInstance.GetMeta())
	}

	// Update the shardInstance
	shardInstance.ProvisionState = metadata.Provision
	if err := datasourceInstance.EnsureExistsShardInstance(db, shardInstance); err != nil {
		t.Fatalf("Error ensuring shardInstance: %v", err)
	}

	// Check
	if !metaEqual(testMeta, datasourceInstance.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, datasourceInstance.GetMeta())
	}

	// Remove all the DBs-- so we can remove the shardInstance
	if err := resetDatasourceInstance(datasourceInstance); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	// Remove the shardInstance
	if err := datasourceInstance.EnsureDoesntExistShardInstance(db.Name, shardInstance.Name); err != nil {
		t.Fatalf("Error EnsureDoesntExistShardInstance: %v", err)
	}

	// TODO: check
}
