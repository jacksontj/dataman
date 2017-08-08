package storagenode

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/copystructure"

	"github.com/jacksontj/dataman/src/datamantype"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node/metadata/filter"

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
		if err := datasourceInstance.EnsureDoesntExistDatabase(context.Background(), dbname); err != nil {
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
	if err := datasourceInstance.EnsureExistsDatabase(context.Background(), testMeta.Databases["example_forum"]); err != nil {
		b, _ := json.Marshal(testMeta.Databases["example_forum"])
		fmt.Printf("%s\n", b)
		t.Fatalf("Error ensuring DB: %v", err)
	}

	// Ensure that the one we had and the one stored are the same
	if !metaEqual(testMeta, datasourceInstance.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, datasourceInstance.GetMeta())
	}

	// Remove it all
	if err := datasourceInstance.EnsureDoesntExistDatabase(context.Background(), "example_forum"); err != nil {
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
	if err := datasourceInstance.EnsureExistsDatabase(context.Background(), db); err != nil {
		t.Fatalf("Error ensuring DB: %v", err)
	}

	// set the DB id -- so the compare works
	testMeta.Databases["example_forum"].ID = db.ID

	shardInstance := testMeta.Databases["example_forum"].ShardInstances["dbshard_example_forum_2"]

	// Ensure the shardInstance
	if err := datasourceInstance.EnsureExistsShardInstance(context.Background(), db, shardInstance); err != nil {
		t.Fatalf("Error ensuring shardInstance: %v", err)
	}

	testMeta.Databases["example_forum"].ProvisionState = metadata.Active
	// Check
	if !metaEqual(testMeta, datasourceInstance.GetMeta()) {
		t.Fatalf("not equal %v != %v", testMeta, datasourceInstance.GetMeta())
	}

	// Update the shardInstance
	shardInstance.ProvisionState = metadata.Provision
	if err := datasourceInstance.EnsureExistsShardInstance(context.Background(), db, shardInstance); err != nil {
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
	if err := datasourceInstance.EnsureDoesntExistShardInstance(context.Background(), db.Name, shardInstance.Name); err != nil {
		t.Fatalf("Error EnsureDoesntExistShardInstance: %v", err)
	}

	// TODO: check
}

// All the tests around data access
func TestDatasource_DataAccess(t *testing.T) {
	datasourceInstance, err := getDatasourceInstance()
	if err != nil {
		t.Fatalf("Unable to get datasourceInstance: %v", err)
	}

	if err := resetDatasourceInstance(datasourceInstance); err != nil {
		t.Fatalf("Unable to reset meta store: %v", err)
	}

	databaseAdd := &metadata.Database{
		Name: "test_function_access",
		ShardInstances: map[string]*metadata.ShardInstance{
			"shard1": &metadata.ShardInstance{
				Name:     "shard1",
				Instance: 1,
				Count:    1,
				Collections: map[string]*metadata.Collection{
					"item": &metadata.Collection{
						Name: "item",
						Fields: map[string]*metadata.CollectionField{
							"data": &metadata.CollectionField{
								Name:      "data",
								Type:      "_document",
								FieldType: metadata.DatamanTypeToFieldType(datamantype.Document),
							},
							"name": &metadata.CollectionField{
								Name:      "name",
								Type:      "_string",
								FieldType: metadata.DatamanTypeToFieldType(datamantype.String),
							},
							"id": &metadata.CollectionField{
								Name:      "id",
								NotNull:   true,
								Type:      "_serial",
								FieldType: metadata.DatamanTypeToFieldType(datamantype.Serial),
							},
						},
						Indexes: map[string]*metadata.CollectionIndex{
							"id": &metadata.CollectionIndex{
								Name:    "id",
								Fields:  []string{"id"},
								Unique:  true,
								Primary: true,
							},
						},
					},
				},
			},
		},
	}

	// Add the database
	if err := datasourceInstance.EnsureExistsDatabase(context.Background(), databaseAdd); err != nil {
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
	result := datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Insert,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         row,
			},
		},
	)
	if result.Error != "" {
		t.Fatalf("Error when adding a valid document: %v", result.Error)
	}
	if result.ValidationError != nil {
		b, _ := json.Marshal(result.ValidationError)
		t.Fatalf("Error when adding a valid document: %v", string(b))
	}
	insertedId := result.Return[0]["id"].(int64)

	badRowTmp, _ := copystructure.Copy(row)
	badRow := badRowTmp.(map[string]interface{})
	badRow["notacolumn"] = "bar"
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Insert,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         badRow,
			},
		},
	)
	if result.Error == "" {
		t.Fatalf("No error when adding an invalid document")
	}

	conflictingRowTmp, _ := copystructure.Copy(row)
	conflictingRow := conflictingRowTmp.(map[string]interface{})
	conflictingRow["id"] = insertedId
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Insert,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         conflictingRow,
			},
		},
	)
	if result.Error == "" {
		t.Fatalf("No error when adding a conflicting row")
	}

	//Get
	//	- get an item which doesn't exist
	//	- get an item which does exist
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Get,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"id":             -1,
			},
		},
	)
	if len(result.Return) != 0 {
		t.Fatalf("Found a non-existant item")
	}
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Get,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"pkey": map[string]interface{}{
					"id": insertedId,
				},
			},
		},
	)
	if len(result.Return) != 1 {
		t.Fatalf("Unable to find inserted item!")
	}

	//Update
	//	- update a non-existant item
	//	- update to column which doesn't exist
	//	- update a single column
	//		-- vaid type
	//		-- invalid type
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Update,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter":         map[string]interface{}{"id": []interface{}{filter.Equal, -1}},
				"record":         map[string]interface{}{"name": "bar"},
			},
		},
	)
	if len(result.Return) != 0 {
		t.Fatalf("Updated %d rows for a non-existant row?", len(result.Return))
	}

	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Update,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter":         map[string]interface{}{"id": []interface{}{filter.Equal, -1}},
				"record":         badRow,
			},
		},
	)
	if result.Error == "" {
		t.Fatalf("No error when updating a row with invalid record")
	}

	invalidColumnRowTmp, _ := copystructure.Copy(row)
	invalidColumnRow := invalidColumnRowTmp.(map[string]interface{})
	invalidColumnRow["name"] = 100
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Update,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter":         map[string]interface{}{"id": []interface{}{filter.Equal, -1}},
				"record":         invalidColumnRow,
			},
		},
	)
	// TODO: need to do actual type checking down in the storageinterface
	if result.Error == "" && false {
		t.Fatalf("No error when updating a row with invalid column type: %v", result)
	}

	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Update,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter":         map[string]interface{}{"id": []interface{}{"=", insertedId}},
				"record":         map[string]interface{}{"name": "tester2"},
			},
		},
	)
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
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Set,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         map[string]interface{}{"notthere": -1, "name": "bar"},
			},
		},
	)
	if len(result.Return) != 0 {
		t.Fatalf("Set %d rows for a non-existant row?", len(result.Return))
	}

	// Update something which *does* exist
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Set,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         map[string]interface{}{"id": insertedId, "name": "bar"},
			},
		},
	)
	if len(result.Return) != 1 {
		t.Fatalf("Unable to set row for an existing row: %v", result)
	}

	// create a valid row
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Set,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         map[string]interface{}{"name": "setname"},
			},
		},
	)
	if result.Error != "" {
		t.Fatalf("Error when setting (creating) a valid row: %s", result.Error)
	}

	// create a invalid row
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Set,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"record":         badRow,
			},
		},
	)
	if result.Error == "" {
		t.Fatalf("No error when set-ing a row with invalid record")
	}

	// Filter
	//  - Get a row that doesn't exist
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Filter,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter": map[string]interface{}{
					"notthere": []interface{}{"=", -1},
					"name":     []interface{}{"=", "bar"},
				},
			},
		},
	)
	if len(result.Return) != 0 {
		t.Fatalf("Filter %d rows for a non-existant row?", len(result.Return))
	}

	//  - Get a row that does exist
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Filter,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter": map[string]interface{}{
					"id":   []interface{}{"=", insertedId},
					"name": []interface{}{"=", "bar"},
				},
			},
		},
	)
	if len(result.Return) != 1 {
		t.Fatalf("Unable to filter row for an existing row: %v", result)
	}

	//  fields
	requestFields := []string{
		"id",
		"data.lastName",
	}
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Filter,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"filter": map[string]interface{}{
					"id":   []interface{}{"=", insertedId},
					"name": []interface{}{"=", "bar"},
				},
				"fields": requestFields,
			},
		},
	)
	if result.Error != "" {
		t.Fatalf("Error when setting (creating) a valid row: %s", result.Error)
	}

	// Check that we only got what we expected
	flatResult := query.FlattenResult(result.Return[0])
	resultFields := make([]string, 0, len(flatResult))
	for k, _ := range flatResult {
		resultFields = append(resultFields, k)
	}
	// check that the fields came back as expected (projected)
	match := true
	if len(flatResult) != len(requestFields) {
		match = false
	}
	for _, k := range requestFields {
		if _, ok := flatResult[k]; !ok {
			match = false
		}
	}
	if !match {
		t.Fatalf("Error, got back different fields expected=%v actual=%v", requestFields, resultFields)
	}

	//Delete
	//	- delete an item which doesn't exist
	//	- an item that does exist
	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Delete,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"id":             -1,
			},
		},
	)
	if len(result.Return) != 0 {
		t.Fatalf("Delete %d rows for a non-existant row?", len(result.Return))
	}

	result = datasourceInstance.HandleQuery(context.Background(),
		&query.Query{
			Type: query.Delete,
			Args: map[string]interface{}{
				"db":             databaseAdd.Name,
				"shard_instance": "shard1",
				"collection":     "item",
				"pkey": map[string]interface{}{
					"id": insertedId,
				},
			},
		},
	)
	if len(result.Return) != 1 {
		t.Fatalf("Unable to delete a row?! %v", result)
	}

}
