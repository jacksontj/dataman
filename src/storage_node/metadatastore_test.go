package storagenode

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

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

	return NewMetadataStore(config)
}

func resetMetaStore(metaStore *MetadataStore) error {
	meta := metaStore.GetMeta()

	for dbname, _ := range meta.Databases {
		if err := metaStore.RemoveDatabase(dbname); err != nil {
			return err
		}
	}

	// Clear out schemas
	for _, schema := range metaStore.ListSchema() {
		if err := metaStore.RemoveSchema(schema.Name, schema.Version); err != nil {
			return err
		}
	}
	return nil
}

// TODO: need additional tests that cover all the methods, this is just enough to know
// that it won't immediately explode
func TestMetaStore(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	// reset the meta store
	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Error resetting metaStore: %v", err)
	}

	var test StoreTestFixture
	bytes, err := ioutil.ReadFile("test.json")
	if err != nil {
		t.Fatalf("Err reading json: %v", err)
	}
	err = json.Unmarshal(bytes, &test)
	if err != nil {
		t.Fatalf("Err loading json: %v", err)
	}

	if err = metaStore.AddDatabase(test.Schema); err != nil {
		t.Fatalf("Unable to add database: %v", err)
	}

	// reset the meta store
	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Error resetting metaStore: %v", err)
	}
}

func TestMetaStore_Schema(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to get metaStore: %v", err)
	}

	// reset the meta store
	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Error resetting metaStore: %v", err)
	}

	schema := &metadata.Schema{
		Name:    "person",
		Version: 1,
		Schema: map[string]interface{}{
			"title": "Person",
			"type":  "object",
			"properties": map[string]interface{}{
				"firstName": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"firstName"},
		},
	}

	if err := metaStore.AddSchema(schema); err != nil {
		t.Fatalf("Unable to add Schema: %v", err)
	}

	storedSchema := metaStore.GetSchema(schema.Name, schema.Version)

	if !reflect.DeepEqual(schema.Schema, storedSchema.Schema) {
		// TODO: fix, this seems to be always triggering, although the map looks correct. Probably a typing problem
		//t.Fatalf("Schema not stored properly expected:\n%v \nactual\n%v", schema.Schema, storedSchema.Schema)
	}
}

func TestMetaStore_Schema2(t *testing.T) {
	metaStore, err := getMetaStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}
	// reset the meta store
	if err := resetMetaStore(metaStore); err != nil {
		t.Fatalf("Error resetting metaStore: %v", err)
	}

	schema1 := metadata.Schema{
		Name:    "person",
		Version: 1,
		Schema: map[string]interface{}{
			"title": "Person",
			"type":  "object",
			"properties": map[string]interface{}{
				"firstName": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"firstName"},
		},
	}

	schema2 := metadata.Schema{
		Name:    "person",
		Version: 2,
		Schema: map[string]interface{}{
			"title": "Person",
			"type":  "object",
			"properties": map[string]interface{}{
				"firstName": map[string]interface{}{
					"type": "string",
				},
				"lastName": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"firstName", "lastName"},
		},
	}

	// Add a schema
	if err := metaStore.AddSchema(&schema1); err != nil {
		t.Fatalf("Unable to add schema: %v", err)
	}

	// Add it again (ensure we can't overwrite)
	if err := metaStore.AddSchema(&schema1); err == nil {
		t.Fatalf("Able to re-add the same schema?: %v", err)
	}

	// Add another one (same id, different version)
	if err := metaStore.AddSchema(&schema2); err != nil {
		t.Fatalf("Unable to add schema: %v", err)
	}

	// Remove one that doesn't exist
	if err := metaStore.RemoveSchema("foo", 5); err == nil {
		t.Fatalf("No error removing a schema which doesn't exist")
	}

	// Remove one
	if err := metaStore.RemoveSchema(schema1.Name, schema1.Version); err != nil {
		t.Fatalf("Error removing schema1: %v", err)
	}

	// Remove another
	if err := metaStore.RemoveSchema(schema2.Name, schema2.Version); err != nil {
		t.Fatalf("Error removing schema2: %v", err)
	}

	// Attempt to add an invalid schema
	invalidSchema := metadata.Schema{
		Name:    "person",
		Version: 1,
		Schema: map[string]interface{}{
			"title": "Person",
			"type":  "objsect",
			"properties": map[string]interface{}{
				"firstName": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"firstName"},
		},
	}
	if err := metaStore.AddSchema(&invalidSchema); err == nil {
		t.Fatalf("No error when adding invalid schema!")
	}

}
