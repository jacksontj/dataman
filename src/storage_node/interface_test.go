package storagenode

// A collection of tests to test the storagenode interface

import (
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

// Test db creation, modification, and removal
func TestDatabase(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}

	// TODO reset meta store

	meta, err := store.GetMeta()
	if err != nil {
		t.Fatalf("Unable to get empty meta from new store: %v", err)
	}

	// Clear the DB -- since we are going to use it
	for _, db := range meta.Databases {
		if err := store.RemoveDatabase(db.Name); err != nil {
			t.Fatalf("Unable to remove DB: %v", err)
		}
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
			},
		},
	}
	tableAdd := &metadata.Table{Name: "table2"}

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

	// Add a table
	if err := store.AddTable(databaseAdd.Name, tableAdd); err != nil {
		t.Fatalf("Error adding table to existing DB: %v", err)
	}
	meta, _ = store.GetMeta()
	if len(meta.Databases[databaseAdd.Name].ListTables()) != 2 {
		t.Fatalf("Error adding table: %v", err)
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
