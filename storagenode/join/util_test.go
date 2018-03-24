package join

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/jacksontj/dataman/storagenode/metadata"
)

func getTestSchema(t *testing.T) *metadata.Meta {
	meta := metadata.NewMeta()
	schemaBytes, err := ioutil.ReadFile("schema.json")
	if err != nil {
		t.Fatalf("Unable to read schema: %v", err)
	}
	if err := json.Unmarshal(schemaBytes, &meta); err != nil {
		t.Fatalf("Unable to load schema: %v", err)
	}
	return meta
}

type joinTestCase struct {
	Collection string // collection to start join on

	Join interface{}

	err bool
}

func Test_OrderNew(t *testing.T) {
	meta := getTestSchema(t)

	db, ok := meta.Databases["test1"]
	if !ok {
		t.Fatalf("Unable to find database: %v", meta.Databases)
	}
	shardInstance, ok := db.ShardInstances["dbshard_test1_9_1"]
	if !ok {
		t.Fatalf("Unable to find shardInstance: %v", db.ShardInstances)
	}

	tests := []joinTestCase{
		{
			Collection: "message",
			Join:       []string{"message.data.thread_ksuid", ".data.thread_ksuid.data.created_by", ".data.created_by"},
		},
		{
			Collection: "message",
			Join:       []string{"thread.data.created_by"},
			err:        true,
		},
		{
			Collection: "thread",
			Join:       []string{"message.data.thread_ksuid", "message.data.thread_ksuid.data.created_by"},
		},
		{
			Collection: "thread",
			Join:       []string{".data.created_by", "message.data.thread_ksuid", "message.data.thread_ksuid.data.created_by"},
		},
		{
			Collection: "user",
			Join:       []string{"thread.data.created_by", "message.data.thread_ksuid"},
		},
		{
			Collection: "user",
			Join:       []string{"thread.data.created_by", "message.data.created_by"},
		},
	}

	getter := func(name string) (MetaCollection, error) {
		return shardInstance.Collections[name], nil
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			joinMap, err := ParseJoinMap(test.Join)
			if err != nil {
				t.Fatalf("Invalid joinMap: %v", err)
			}
			c, err := OrderJoins(getter, shardInstance.Collections[test.Collection], joinMap)
			if test.err != (err != nil) {
				if test.err {
					t.Fatalf("No Error on %d when expected", i)
				} else {
					t.Fatalf("Error on %d when not expected: %v", i, err)
				}
			}
			if err == nil {
				cBytes, _ := json.MarshalIndent(c, "", "  ")

				ioutil.WriteFile("/tmp/output", cBytes, 0644)
			}
		})
	}

}
