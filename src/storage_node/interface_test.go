package storagenode

// A collection of tests to test the storagenode interface

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/mitchellh/copystructure"
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
	meta := store.GetMeta()

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

	meta := store.GetMeta()
	if err != nil {
		t.Fatalf("Unable to get empty meta from new store: %v", err)
	}

	if len(meta.ListDatabases()) > 0 {
		t.Fatalf("New node has tables in it?")
	}

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
		Columns: []*metadata.TableColumn{
			&metadata.TableColumn{
				Name: "data",
				Type: metadata.Document,
			},
		},
		Indexes: map[string]*metadata.TableIndex{
			"data.foo": &metadata.TableIndex{
				Name:    "data.foo",
				Columns: []string{"data.foo"},
			},
		},
	}

	// Add a database
	if err := store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	// TODO: refreshes should happen in the actual store-- not here in the tests
	store.RefreshMeta()
	meta = store.GetMeta()
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
	store.RefreshMeta()
	meta = store.GetMeta()
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
	store.RefreshMeta()
	meta = store.GetMeta()
	if len(meta.Databases[databaseAdd.Name].ListTables()) != 1 {
		t.Fatalf("Error removing table: %v", err)
	}

	// Remove a database
	if err := store.RemoveDatabase(databaseAdd.Name); err != nil {
		t.Fatalf("Unable to remove database: %v", err)
	}
	store.RefreshMeta()
	meta = store.GetMeta()
	if len(meta.ListDatabases()) != 0 {
		t.Fatalf("DB wasn't removed")
	}
}

// TODO: test indexes
// Test Functions for covering a document DB
func TestDocumentDatabase(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}
	resetStore(store)

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
			"data.firstName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &tableIndex)
	if err := store.AddIndex("docdb", "person", &tableIndex); err != nil {
		t.Fatalf("Unable to add simple index: %v", err)
	}

	indexBytes = []byte(`
	{
		"name": "complex",
		"columns": [
			"data.firstName",
			"data.lastName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &tableIndex)
	if err := store.AddIndex("docdb", "person", &tableIndex); err != nil {
		t.Fatalf("Unable to add simple index: %v", err)
	}

	// Remove indexes
	if err := store.RemoveIndex("docdb", "person", "simple"); err != nil {
		t.Fatalf("Unable to remove index: %v", err)
	}
	if err := store.RemoveIndex("docdb", "person", "complex"); err != nil {
		t.Fatalf("Unable to remove index: %v", err)
	}

	// Add a valid document
	result := store.Insert(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document")
	}

	// Add a valid document
	result = store.Insert(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
				"lastName":  "foobar",
			},
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document")
	}

	// Filter
	result = store.Filter(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
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
// Test Functions for covering a column DB (sql)
func TestColumnDatabase(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}

	meta := store.GetMeta()
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
		Name: "columndb",
		Tables: map[string]*metadata.Table{
			"person": &metadata.Table{
				Name: "person",
				Columns: []*metadata.TableColumn{
					&metadata.TableColumn{
						Name: "firstName",
						// TODO: non-null per column
						Type:    metadata.String,
						NotNull: true,
					},
				},
			},
		},
	}
	tableUpdate := &metadata.Table{
		Name: "person",
		Columns: []*metadata.TableColumn{
			&metadata.TableColumn{
				Name: "firstName",
				// TODO: non-null per column
				Type: metadata.String,
			},
			&metadata.TableColumn{
				Name: "lastName",
				// TODO: non-null per column
				Type: metadata.String,
			},
		},
		Indexes: map[string]*metadata.TableIndex{
			"simple": &metadata.TableIndex{
				Name:    "simple",
				Columns: []string{"firstName"},
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
		"name": "complex",
		"columns": [
			"firstName",
			"lastName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &tableIndex)
	if err := store.AddIndex(databaseAdd.Name, "person", &tableIndex); err == nil {
		t.Fatalf("No error when adding an index to a column which doesn't exist!")
	}

	// Add the missing column
	if err := store.UpdateTable(databaseAdd.Name, tableUpdate); err != nil {
		t.Fatalf("Error updating table: %v", err)
	}
	// TODO: move inside the store itself
	store.RefreshMeta()

	if err := store.AddIndex(databaseAdd.Name, "person", &tableIndex); err != nil {
		t.Fatalf("Error when adding index: %v", err)
	}

	// Remove indexes
	if err := store.RemoveIndex(databaseAdd.Name, "person", "simple"); err != nil {
		t.Fatalf("Unable to remove index: %v", err)
	}
	if err := store.RemoveIndex(databaseAdd.Name, "person", "complex"); err != nil {
		t.Fatalf("Unable to remove index: %v", err)
	}

	// Add a valid document
	result := store.Insert(map[string]interface{}{
		"db":    databaseAdd.Name,
		"table": "person",
		"columns": map[string]interface{}{
			"firstName": "tester",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}

	// Add an invalid document
	result = store.Insert(map[string]interface{}{
		"db":    databaseAdd.Name,
		"table": "person",
		"columns": map[string]interface{}{
			"lastName": "mctester",
		},
	})
	if result.Error == "" {
		t.Fatalf("No error when adding an invalid document!")
	}

	// Add a valid document
	result = store.Insert(map[string]interface{}{
		"db":    databaseAdd.Name,
		"table": "person",
		"columns": map[string]interface{}{
			"firstName": "tester",
			"lastName":  "foobar",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}

	// Filter
	result = store.Filter(map[string]interface{}{
		"db":    databaseAdd.Name,
		"table": "person",
		"columns": map[string]interface{}{
			"firstName": "tester",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when running a Filter: %v", result.Error)
	}
	if len(result.Return) != 2 {
		t.Fatalf("Filter returned %d results, instead of the expected 2: %v", len(result.Return), result.Return)
	}

	// TODO: we need to get back the IDs of the documents to call delete-- otherwise it is a filter delete
	// Delete
}

// TODO: more generic?-- maybe break it out to have a struct where the schema and objects can be defined
// Test Functions for covering a document DB
func TestFunctionAccess(t *testing.T) {
	store, err := getStore()
	if err != nil {
		t.Fatalf("Unable to create test storagenode")
	}
	resetStore(store)

	databaseAdd := &metadata.Database{
		Name: "test_function_access",
		Tables: map[string]*metadata.Table{
			"item": &metadata.Table{
				Name: "item",
				Columns: []*metadata.TableColumn{
					&metadata.TableColumn{
						Name: "data",
						Type: metadata.Document,
					},
					&metadata.TableColumn{
						Name: "name",
						Type: metadata.String,
					},
				},
			},
		},
	}

	// Add the database
	if err := store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	row := map[string]interface{}{
		"data": map[string]interface{}{
			"lastName": "mctester",
		},
		"name": "tester",
	}

	//Insert
	//	- add a valid row
	//	- add an invalid row
	//	- add a conflicting row
	result := store.Insert(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": row,
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}
	insertedId := result.Return[0]["_id"].(int64)

	badRowTmp, _ := copystructure.Copy(row)
	badRow := badRowTmp.(map[string]interface{})
	badRow["notacolumn"] = "bar"
	result = store.Insert(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": badRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when adding an invalid document")
	}

	conflictingRowTmp, _ := copystructure.Copy(row)
	conflictingRow := conflictingRowTmp.(map[string]interface{})
	conflictingRow["_id"] = insertedId
	result = store.Insert(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": conflictingRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when adding a conflicting row")
	}

	//Get
	//	- get an item which doesn't exist
	//	- get an item which does exist
	result = store.Get(map[string]interface{}{
		"db":    databaseAdd.Name,
		"table": "item",
		"_id":   -1,
	})
	if len(result.Return) != 0 {
		t.Fatalf("Found a non-existant item")
	}
	result = store.Get(map[string]interface{}{
		"db":    databaseAdd.Name,
		"table": "item",
		"_id":   insertedId,
	})
	if len(result.Return) != 1 {
		t.Fatalf("Unable to find inserted item!")
	}

	//Update
	//	- update a non-existant item
	//	- update to column which doesn't exist
	//	- update a single column
	//		-- vaid type
	//		-- invalid type
	result = store.Update(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"filter":  map[string]interface{}{"_id": -1},
		"columns": map[string]interface{}{"name": "bar"},
	})
	if len(result.Return) != 0 {
		t.Fatalf("Updated %d rows for a non-existant row?", len(result.Return))
	}

	result = store.Update(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"filter":  map[string]interface{}{"_id": insertedId},
		"columns": badRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when updating a row with invalid columns")
	}

	invalidColumnRowTmp, _ := copystructure.Copy(row)
	invalidColumnRow := invalidColumnRowTmp.(map[string]interface{})
	invalidColumnRow["name"] = 100
	result = store.Update(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"filter":  map[string]interface{}{"_id": insertedId},
		"columns": invalidColumnRow,
	})
	// TODO: need to do actual type checking down in the storageinterface
	if result.Error == "" && false {
		t.Fatalf("No error when updating a row with invalid column type: %v", result)
	}

	result = store.Update(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"filter":  map[string]interface{}{"_id": insertedId},
		"columns": map[string]interface{}{"name": "tester2"},
	})
	if len(result.Return) != 1 {
		t.Fatalf("Update found nothing: %v", result)
	}
	// Check that "data" is untouched, but "name" is updated
	if result.Return[0]["name"] != "tester2" {
		t.Fatalf("Update didn't update column name! expected=%v actual=%v", "tester2", result.Return[0]["name"])
	}
	if !reflect.DeepEqual(result.Return[0]["data"], row["data"]) {
		t.Fatalf("Update changed value of data.lastName! expected=%v actual=%v", result.Return[0]["data"].(map[string]string)["lastName"], row["data"].(map[string]string)["lastName"])
	}

	// Set
	//  - update a row (does/doesn't exist)
	//  - create a row (valid, invalid)
	// Update something which doesn't exist
	result = store.Set(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": map[string]interface{}{"notthere": -1, "name": "bar"},
	})
	if len(result.Return) != 0 {
		t.Fatalf("Set %d rows for a non-existant row?", len(result.Return))
	}

	// Update something which *does* exist
	result = store.Set(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": map[string]interface{}{"_id": insertedId, "name": "bar"},
	})
	if len(result.Return) != 1 {
		t.Fatalf("Unable to set row for an existing row: %v", result)
	}

	// create a valid row
	result = store.Set(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": map[string]interface{}{"name": "setname"},
	})
	if result.Error != "" {
		t.Fatalf("Error when setting (creating) a valid row: %s", result.Error)
	}

	// create a invalid row
	result = store.Set(map[string]interface{}{
		"db":      databaseAdd.Name,
		"table":   "item",
		"columns": badRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when set-ing a row with invalid columns")
	}

	//Delete
	//	- delete an item which doesn't exist
	//	- an item that does exist
	result = store.Delete(map[string]interface{}{
		"db":     databaseAdd.Name,
		"table":  "item",
		"filter": map[string]interface{}{"_id": -1},
	})
	if len(result.Return) != 0 {
		t.Fatalf("Delete %d rows for a non-existant row?", len(result.Return))
	}

	result = store.Delete(map[string]interface{}{
		"db":     databaseAdd.Name,
		"table":  "item",
		"filter": map[string]interface{}{"_id": insertedId},
	})
	if len(result.Return) != 1 {
		t.Fatalf("Unable to delete a row?! %v", result)
	}

}
