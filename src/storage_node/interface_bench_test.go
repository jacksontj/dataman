package storagenode

// A collection of benchmarks to test the storagenode interface

import (
	"encoding/json"
	"testing"

	"github.com/jacksontj/dataman/src/metadata"
)

func BenchmarkDocumentDatabase(b *testing.B) {
	store, err := getStore()
	if err != nil {
		b.Fatalf("Unable to create test storagenode")
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
		b.Fatalf("Error adding database: %v", err)
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
		b.Fatalf("Unable to add simple index: %v", err)
	}

	// Insert single item
	result := store.Set(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
		},
	})
	if result.Error != "" {
		b.Fatalf("Error when adding a valid document")
	}

	b.Run("Get", func(b *testing.B) { benchDocument_Get(b, store) })
	b.Run("Set", func(b *testing.B) { benchDocument_Set(b, store) })
	b.Run("Delete", func(b *testing.B) { benchDocument_Delete(b, store) })
	b.Run("Filter", func(b *testing.B) { benchDocument_Filter(b, store) })

}

func benchDocument_Get(b *testing.B, store StorageInterface) {
	// Filter
	result := store.Filter(map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
		},
	})
	if result.Error != "" {
		b.Fatalf("Error when adding a valid document")
	}

	id, ok := result.Return[0]["_id"]
	if !ok {
		b.Fatalf("Unable to get _id")
	}

	query := map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"_id":   id,
	}

	// Initialization done, lets do some benchmarking
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		store.Get(query)
	}
}

func benchDocument_Set(b *testing.B, store StorageInterface) {
	// Insert single item
	query := map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
		},
	}
	// Initialization done, lets do some benchmarking
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		query["columns"].(map[string]interface{})["data"].(map[string]interface{})["firstName"] = n
		store.Set(query)
	}
}

func benchDocument_Delete(b *testing.B, store StorageInterface) {
	// Insert single item
	query := map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
		},
	}
	// Insert N items
	for n := 0; n < b.N; n++ {
		query["columns"].(map[string]interface{})["data"].(map[string]interface{})["firstName"] = n
		store.Set(query)
	}

	// Initialization done, lets do some benchmarking
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		query["columns"].(map[string]interface{})["data"].(map[string]interface{})["firstName"] = n
		store.Delete(query)
	}
}

func benchDocument_Filter(b *testing.B, store StorageInterface) {
	// Insert single item
	query := map[string]interface{}{
		"db":    "docdb",
		"table": "person",
		"columns": map[string]interface{}{
			"data": map[string]interface{}{
				"firstName": "tester",
			},
		},
	}
	// Insert N items
	// TODO: vary the number of items we are getting in the filter?
	for n := 0; n < 10; n++ {
		query["columns"].(map[string]interface{})["data"].(map[string]interface{})["firstName"] = n
		store.Set(query)
	}

	query = map[string]interface{}{
		"db":    "docdb",
		"table": "person",
	}

	// Initialization done, lets do some benchmarking
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		store.Filter(query)
	}
}

/*
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
	result := store.Set(map[string]interface{}{
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
	result = store.Set(map[string]interface{}{
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
	result := store.Set(map[string]interface{}{
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
	result = store.Set(map[string]interface{}{
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
	result = store.Set(map[string]interface{}{
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
*/
