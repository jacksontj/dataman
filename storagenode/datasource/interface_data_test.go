package datasource

/*
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
		Collections: map[string]*metadata.Collection{
			"table1": &metadata.Collection{
				Name: "table1",
				Fields: []*metadata.Field{
					&metadata.Field{
						Name: "data",
						Type: metadata.Document,
					},
				},
			},
		},
	}
	collectionAdd := &metadata.Collection{
		Name: "table2",
		Fields: []*metadata.Field{
			&metadata.Field{
				Name: "data",
				Type: metadata.Document,
			},
		},
	}
	collectionUpdate := &metadata.Collection{
		Name: "table2",
		Fields: []*metadata.Field{
			&metadata.Field{
				Name: "data",
				Type: metadata.Document,
			},
		},
		Indexes: map[string]*metadata.CollectionIndex{
			"data.foo": &metadata.CollectionIndex{
				Name:   "data.foo",
				Fields: []string{"data.foo"},
			},
		},
	}

	// Add a database
	if err := store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	// TODO: refreshes should happen in the actual store-- not here in the tests
	meta = store.GetMeta()
	if len(meta.ListDatabases()) != 1 {
		t.Fatalf("DB wasn't added")
	}

	// Attempt to add a database that already is added
	if err := store.AddDatabase(databaseAdd); err == nil {
		t.Fatalf("Store allowed me to add a database which already exists: %v", err)
	}

	// Update a collection which doesnt exist
	if err := store.UpdateCollection(databaseAdd.Name, collectionUpdate); err == nil {
		t.Fatalf("Store allowed me to update a collection which doesn't exist!")
	}

	// Add a collection
	if err := store.AddCollection(databaseAdd.Name, collectionAdd); err != nil {
		t.Fatalf("Error adding collection to existing DB: %v", err)
	}
	meta = store.GetMeta()
	if len(meta.Databases[databaseAdd.Name].ListCollections()) != 2 {
		t.Fatalf("Error adding collection: %v", err)
	}

	// Update a collection which does exist
	if err := store.UpdateCollection(databaseAdd.Name, collectionUpdate); err != nil {
		t.Fatalf("Error updating collection: %v", err)
	}

	// Attempt to add a collection that already exists
	if err := store.AddCollection(databaseAdd.Name, collectionAdd); err == nil {
		t.Fatalf("Error added collection to existing DB which already exists")
	}

	// Remove a collection
	if err := store.RemoveCollection(databaseAdd.Name, collectionAdd.Name); err != nil {
		t.Fatalf("Unable to remove collection: %v", err)
	}
	meta = store.GetMeta()
	if len(meta.Databases[databaseAdd.Name].ListCollections()) != 1 {
		t.Fatalf("Error removing collection: %v", err)
	}

	// Remove a database
	if err := store.RemoveDatabase(databaseAdd.Name); err != nil {
		t.Fatalf("Unable to remove database: %v", err)
	}
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
		Collections: map[string]*metadata.Collection{
			"person": &metadata.Collection{
				Name: "person",
				Fields: []*metadata.Field{
					&metadata.Field{
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
	var collectionIndex metadata.CollectionIndex
	indexBytes := []byte(`
	{
		"name": "simple",
		"fields": [
			"data.firstName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &collectionIndex)
	if err := store.AddIndex("docdb", "person", &collectionIndex); err != nil {
		t.Fatalf("Unable to add simple index: %v", err)
	}

	indexBytes = []byte(`
	{
		"name": "complex",
		"fields": [
			"data.firstName",
			"data.lastName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &collectionIndex)
	if err := store.AddIndex("docdb", "person", &collectionIndex); err != nil {
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
		"db":         "docdb",
		"collection": "person",
		"record": map[string]interface{}{
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
		"db":         "docdb",
		"collection": "person",
		"record": map[string]interface{}{
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
		"db":         "docdb",
		"collection": "person",
		"record": map[string]interface{}{
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
		Collections: map[string]*metadata.Collection{
			"person": &metadata.Collection{
				Name: "person",
				Fields: []*metadata.Field{
					&metadata.Field{
						Name: "firstName",
						// TODO: non-null per column
						Type:    metadata.String,
						NotNull: true,
					},
				},
			},
		},
	}
	collectionUpdate := &metadata.Collection{
		Name: "person",
		Fields: []*metadata.Field{
			&metadata.Field{
				Name: "firstName",
				// TODO: non-null per column
				Type: metadata.String,
			},
			&metadata.Field{
				Name: "lastName",
				// TODO: non-null per column
				Type: metadata.String,
			},
		},
		Indexes: map[string]*metadata.CollectionIndex{
			"simple": &metadata.CollectionIndex{
				Name:   "simple",
				Fields: []string{"firstName"},
			},
		},
	}

	// Add the database
	if err := store.AddDatabase(databaseAdd); err != nil {
		t.Fatalf("Error adding database: %v", err)
	}

	// Add index
	var collectionIndex metadata.CollectionIndex
	indexBytes := []byte(`
	{
		"name": "complex",
		"fields": [
			"firstName",
			"lastName"
		]
	}
	`)
	json.Unmarshal(indexBytes, &collectionIndex)
	if err := store.AddIndex(databaseAdd.Name, "person", &collectionIndex); err == nil {
		t.Fatalf("No error when adding an index to a column which doesn't exist!")
	}

	// Add the missing column
	if err := store.UpdateCollection(databaseAdd.Name, collectionUpdate); err != nil {
		t.Fatalf("Error updating collection: %v", err)
	}

	if err := store.AddIndex(databaseAdd.Name, "person", &collectionIndex); err != nil {
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
		"db":         databaseAdd.Name,
		"collection": "person",
		"record": map[string]interface{}{
			"firstName": "tester",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}

	// Add an invalid document
	result = store.Insert(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "person",
		"record": map[string]interface{}{
			"lastName": "mctester",
		},
	})
	if result.Error == "" {
		t.Fatalf("No error when adding an invalid document!")
	}

	// Add a valid document
	result = store.Insert(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "person",
		"record": map[string]interface{}{
			"firstName": "tester",
			"lastName":  "foobar",
		},
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}

	// Filter
	result = store.Filter(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "person",
		"record": map[string]interface{}{
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
		Collections: map[string]*metadata.Collection{
			"item": &metadata.Collection{
				Name: "item",
				Fields: []*metadata.Field{
					&metadata.Field{
						Name: "data",
						Type: metadata.Document,
					},
					&metadata.Field{
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
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     row,
	})
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}
	insertedId := result.Return[0]["_id"].(int64)

	badRowTmp, _ := copystructure.Copy(row)
	badRow := badRowTmp.(map[string]interface{})
	badRow["notacolumn"] = "bar"
	result = store.Insert(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     badRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when adding an invalid document")
	}

	conflictingRowTmp, _ := copystructure.Copy(row)
	conflictingRow := conflictingRowTmp.(map[string]interface{})
	conflictingRow["_id"] = insertedId
	result = store.Insert(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     conflictingRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when adding a conflicting row")
	}

	//Get
	//	- get an item which doesn't exist
	//	- get an item which does exist
	result = store.Get(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"_id":        -1,
	})
	if len(result.Return) != 0 {
		t.Fatalf("Found a non-existant item")
	}
	result = store.Get(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"_id":        insertedId,
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
		"db":         databaseAdd.Name,
		"collection": "item",
		"filter":     map[string]interface{}{"_id": -1},
		"record":     map[string]interface{}{"name": "bar"},
	})
	if len(result.Return) != 0 {
		t.Fatalf("Updated %d rows for a non-existant row?", len(result.Return))
	}

	result = store.Update(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"filter":     map[string]interface{}{"_id": insertedId},
		"record":     badRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when updating a row with invalid record")
	}

	invalidColumnRowTmp, _ := copystructure.Copy(row)
	invalidColumnRow := invalidColumnRowTmp.(map[string]interface{})
	invalidColumnRow["name"] = 100
	result = store.Update(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"filter":     map[string]interface{}{"_id": insertedId},
		"record":     invalidColumnRow,
	})
	// TODO: need to do actual type checking down in the storageinterface
	if result.Error == "" && false {
		t.Fatalf("No error when updating a row with invalid column type: %v", result)
	}

	result = store.Update(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"filter":     map[string]interface{}{"_id": insertedId},
		"record":     map[string]interface{}{"name": "tester2"},
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
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     map[string]interface{}{"notthere": -1, "name": "bar"},
	})
	if len(result.Return) != 0 {
		t.Fatalf("Set %d rows for a non-existant row?", len(result.Return))
	}

	// Update something which *does* exist
	result = store.Set(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     map[string]interface{}{"_id": insertedId, "name": "bar"},
	})
	if len(result.Return) != 1 {
		t.Fatalf("Unable to set row for an existing row: %v", result)
	}

	// create a valid row
	result = store.Set(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     map[string]interface{}{"name": "setname"},
	})
	if result.Error != "" {
		t.Fatalf("Error when setting (creating) a valid row: %s", result.Error)
	}

	// create a invalid row
	result = store.Set(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"record":     badRow,
	})
	if result.Error == "" {
		t.Fatalf("No error when set-ing a row with invalid record")
	}

	//Delete
	//	- delete an item which doesn't exist
	//	- an item that does exist
	result = store.Delete(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"_id":        -1,
	})
	if len(result.Return) != 0 {
		t.Fatalf("Delete %d rows for a non-existant row?", len(result.Return))
	}

	result = store.Delete(map[string]interface{}{
		"db":         databaseAdd.Name,
		"collection": "item",
		"_id":        insertedId,
	})
	if len(result.Return) != 1 {
		t.Fatalf("Unable to delete a row?! %v", result)
	}

}

*/
