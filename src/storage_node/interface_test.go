package storagenode

// A collection of tests to test the storagenode interface

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/jacksontj/dataman/src/metadata"
)

// TODO: have a list of them? We want to test all of them (or become a library of tests
// that the modules can just run
func getStore() (StorageInterface, error) {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("storagenode/config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		return nil, err
	}

	// Load the store we are responsible for
	store, err := config.GetStore()
	if err != nil {
		return nil, err
	}

	return store, nil
}

func resetStore(store StorageInterface) error {
	meta, err := store.GetMeta()
	if err != nil {
		return err
	}

	// Clear the DB -- since we are going to use it
	for _, db := range meta.Databases {
		if err := store.RemoveDatabase(db.Name); err != nil {
			return err
		}
	}

	// Validate that the schemas we want to add aren't there (remove if they are)
	schemas := store.ListSchemas()
	for _, schema := range schemas {
		if err := store.RemoveSchema(schema.Name, schema.Version); err != nil {
			return err
		}
	}
	return nil
}

// Test db creation, modification, and removal
func TestSchema(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}
	resetStore(store)

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
	if err := store.AddSchema(&schema1); err != nil {
		t.Fatalf("Unable to add schema: %v", err)
	}

	// Add it again (ensure we can't overwrite)
	if err := store.AddSchema(&schema1); err == nil {
		t.Fatalf("Able to re-add the same schema?: %v", err)
	}

	// Add another one (same id, different version)
	if err := store.AddSchema(&schema2); err != nil {
		t.Fatalf("Unable to add schema: %v", err)
	}

	// Remove one that doesn't exist
	if err := store.RemoveSchema("foo", 5); err == nil {
		t.Fatalf("No error removing a schema which doesn't exist")
	}

	// Remove one
	if err := store.RemoveSchema(schema1.Name, schema1.Version); err != nil {
		t.Fatalf("Error removing schema1: %v", err)
	}

	// Remove another
	if err := store.RemoveSchema(schema2.Name, schema2.Version); err != nil {
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
	if err := store.AddSchema(&invalidSchema); err == nil {
		t.Fatalf("No error when adding invalid schema!")
	}

}

// Test db creation, modification, and removal
func TestDatabase(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}

	resetStore(store)

	// TODO reset meta store

	meta, err := store.GetMeta()
	if err != nil {
		t.Fatalf("Unable to get empty meta from new store: %v", err)
	}

	if len(meta.ListDatabases()) > 0 {
		t.Fatalf("New node has tables in it?")
	}

	// TODO: add document schema tests
	// TODO: add index tests

	databaseAdd := &metadata.Database{
		Name: "testdb",
		Tables: map[string]*metadata.Table{
			"table1": &metadata.Table{
				Name: "table1",
				Columns: []*metadata.TableColumn{
					&metadata.TableColumn{
						Name: "data",
						Type: metadata.Document,
					},
				},
			},
		},
	}
	tableAdd := &metadata.Table{
		Name: "table2",
		Columns: []*metadata.TableColumn{
			&metadata.TableColumn{
				Name: "data",
				Type: metadata.Document,
			},
		},
	}
	tableUpdate := &metadata.Table{
		Name: "table2",
		Indexes: map[string]*metadata.TableIndex{
			"foo": &metadata.TableIndex{
				Name:    "foo",
				Columns: []string{"foo"},
			},
		},
	}

	// Add a database
	if err := store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	meta, _ = store.GetMeta()
	if len(meta.ListDatabases()) != 1 {
		t.Fatalf("DB wasn't added")
	}

	// Attempt to add a database that already is added
	if err := store.AddDatabase(databaseAdd); err == nil {
		t.Fatalf("Store allowed me to add a database which already exists: %v", err)
	}

	// Update a table which doesnt exist
	if err := store.UpdateTable(databaseAdd.Name, tableUpdate); err == nil {
		t.Fatalf("Store allowed me to update a table which doesn't exist!")
	}

	// Add a table
	if err := store.AddTable(databaseAdd.Name, tableAdd); err != nil {
		t.Fatalf("Error adding table to existing DB: %v", err)
	}
	meta, _ = store.GetMeta()
	if len(meta.Databases[databaseAdd.Name].ListTables()) != 2 {
		t.Fatalf("Error adding table: %v", err)
	}

	// Update a table which does exist
	if err := store.UpdateTable(databaseAdd.Name, tableUpdate); err != nil {
		t.Fatalf("Error updating table: %v", err)
	}

	// Attempt to add a table that already exists
	if err := store.AddTable(databaseAdd.Name, tableAdd); err == nil {
		t.Fatalf("Error added table to existing DB which already exists")
	}

	// Remove a table
	if err := store.RemoveTable(databaseAdd.Name, tableAdd.Name); err != nil {
		t.Fatalf("Unable to remove table: %v", err)
	}
	meta, _ = store.GetMeta()
	if len(meta.Databases[databaseAdd.Name].ListTables()) != 1 {
		t.Fatalf("Error removing table: %v", err)
	}

	// Remove a database
	if err := store.RemoveDatabase(databaseAdd.Name); err != nil {
		t.Fatalf("Unable to remove database: %v", err)
	}
	meta, _ = store.GetMeta()
	if len(meta.ListDatabases()) != 0 {
		t.Fatalf("DB wasn't removed")
	}
}

// Test Functions for covering a document DB
func TestDocumentDatabase(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}

	meta, err := store.GetMeta()
	if err != nil {
		t.Fatalf("Unable to get empty meta from new store: %v", err)
	}

	// TODO: move into getStore()
	// Clear the DB -- since we are going to use it
	for _, db := range meta.Databases {
		if err := store.RemoveDatabase(db.Name); err != nil {
			t.Fatalf("Unable to remove DB: %v", err)
		}
	}

	// TODO: add document schema tests
	// TODO: add index tests

	databaseAdd := &metadata.Database{
		Name: "docdb",
		Tables: map[string]*metadata.Table{
			"person": &metadata.Table{
				Name: "person",
				Columns: []*metadata.TableColumn{
					&metadata.TableColumn{
						Name: "data",
						Type: metadata.Document,
						Schema: &metadata.Schema{
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
						},
					},
				},
			},
		},
	}

	// Add the database
	if err := store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	// Add index
	var tableIndex metadata.TableIndex
	indexBytes := []byte(`
	{
		"name": "simple",
		"columns": [
			"firstName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &tableIndex)
	if err := store.AddIndex("docdb", "person", &tableIndex); err != nil {
		t.Fatalf("Unable to add simple index")
	}

	indexBytes = []byte(`
	{
		"name": "complex",
		"columns": [
			"firstName",
			"lastName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &tableIndex)
	if err := store.AddIndex("docdb", "person", &tableIndex); err != nil {
		t.Fatalf("Unable to add simple index")
	}

	// Remove indexes
	if err := store.RemoveIndex("docdb", "person", "simple"); err != nil {
		t.Fatalf("Unable to remove index: %v", err)
	}
	if err := store.RemoveIndex("docdb", "person", "complex"); err != nil {
		t.Fatalf("Unable to remove index: %v", err)
	}

	// Add a valid document
	result := store.Set(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"data": map[string]interface{}{
			"fistName": "tester",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document")
	}

	// Add a valid document
	result = store.Set(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"data": map[string]interface{}{
			"fistName": "tester",
			"lastName": "foobar",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document")
	}

	// Filter
	result = store.Filter(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"data": map[string]interface{}{
			"fistName": "tester",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document")
	}
	if len(result.Return) != 2 {
		t.Fatalf("Filter returned %d results, instead of the expected 2: %v", len(result.Return), result.Return)
	}

	// TODO: we need to get back the IDs of the documents to call delete-- otherwise it is a filter delete
	// Delete
}

// TODO: test indexes
