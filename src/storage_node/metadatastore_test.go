package storagenode

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"

	"github.com/jacksontj/dataman/src/storage_node/metadata"

	"gopkg.in/yaml.v2"
)

func getMetaStore() (*MetadataStore, error) {
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

	return NewMetadataStore(datasourceInstanceConfig)
}

func resetMetaStore(metaStore *MetadataStore) error {
	meta := metaStore.GetMeta()

	for dbname, _ := range meta.Databases {
		if err := metaStore.EnsureDoesntExistDatabase(dbname); err != nil {
			return err
		}
	}

	return nil
}

// We have a variety of smaller internal fields which we don't care about for
// the use of comparison. So we'll just json dump and compare
func metaEqual(a, b interface{}) bool {
	aBytes, _ := json.MarshalIndent(a, "", "  ")
	bBytes, _ := json.MarshalIndent(b, "", "  ")

	ioutil.WriteFile("/home/jacksontj/Desktop/tmp/a", aBytes, 0644)
	ioutil.WriteFile("/home/jacksontj/Desktop/tmp/b", bBytes, 0644)

	if len(aBytes) != len(bBytes) {
		return false
	}

	for i, b := range aBytes {
		if b != bBytes[i] {
			return false
		}
	}
	return true
}

func getTestMeta() (*metadata.Meta, error) {
	testMeta := &metadata.Meta{}
	metaString, err := ioutil.ReadFile("test_metadata.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(metaString), &testMeta); err != nil {
		return nil, err
	}
	return testMeta, nil
}

func TestMetaStore_Database(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	testMeta, err := getTestMeta()
	if err != nil {
		logrus.Fatalf("Error loading test meta: %v", err)
	}

	// Insert the meta -- here the provision state is all 0
	if err := metaStore.EnsureExistsDatabase(testMeta.Databases["example_forum"]); err != nil {
		t.Fatalf("Error ensuring DB: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(testMeta, metaStore.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, metaStore.GetMeta())
	}

	// Now lets update the provision state for stuff
	db := metaStore.GetMeta().Databases["example_forum"]
	db.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsDatabase(db); err != nil {
		t.Fatalf("Error ensuring DB 2: %v", err)
	}

	// Make sure it changed
	if !metaEqual(db, metaStore.GetMeta().Databases["example_forum"]) {
		t.Fatalf("not equal %v != %v", testMeta, metaStore.GetMeta())
	}

	// Remove it all
	if err := metaStore.EnsureDoesntExistDatabase("example_forum"); err != nil {
		t.Fatalf("Error EnsureDoesntExistDatabase: %v", err)
	}

	// TODO: check
}

func TestMetaStore_ShardInstance(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	testMeta, err := getTestMeta()
	if err != nil {
		logrus.Fatalf("Error loading test meta: %v", err)
	}

	db := &metadata.Database{Name: "example_forum"}
	// Insert the db
	if err := metaStore.EnsureExistsDatabase(db); err != nil {
		t.Fatalf("Error ensuring DB: %v", err)
	}

	// set the DB id -- so the compare works
	testMeta.Databases["example_forum"].ID = db.ID

	shardInstance := testMeta.Databases["example_forum"].ShardInstances["dbshard_example_forum_2"]

	// Ensure the shardInstance
	if err := metaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		t.Fatalf("Error ensuring shardInstance: %v", err)
	}

	// Check
	if !metaEqual(testMeta, metaStore.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, metaStore.GetMeta())
	}

	// Update the shardInstance
	shardInstance.ProvisionState = metadata.Provision
	if err := metaStore.EnsureExistsShardInstance(db, shardInstance); err != nil {
		t.Fatalf("Error ensuring shardInstance: %v", err)
	}

	// Check
	if !metaEqual(testMeta, metaStore.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, metaStore.GetMeta())
	}

	// Remove the shardInstance
	if err := metaStore.EnsureDoesntExistShardInstance(db.Name, shardInstance.Name); err != nil {
		t.Fatalf("Error EnsureDoesntExistShardInstance: %v", err)
	}

	// TODO: check
}
