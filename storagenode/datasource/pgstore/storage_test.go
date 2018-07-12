package pgstorage

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

var filterTestCases []*filterTestCase

func init() {
	filterTestCases = []*filterTestCase{
		{
			filter: map[string]interface{}{"id": []interface{}{"=", 1}},
			result: ` id='1'`,
		},

		{
			filter: "AND",
			err:    true,
		},

		{
			filter: []interface{}{
				map[string]interface{}{"id": []interface{}{"=", 1}},
				"OR",
				map[string]interface{}{"id": []interface{}{"=", 2}},
			},
			result: `( id='1' OR  id='2')`,
		},

		{
			filter: map[string]interface{}{},
			result: `true`,
		},

		{
			filter: []interface{}{
				map[string]interface{}{},
				"OR",
				map[string]interface{}{"id": []interface{}{"=", 2}},
			},
			result: `(true OR  id='2')`,
		},
		/*
			{
				filter: map[string]interface{}{"jsonb.bar": []interface{}{"=", 1}},
				result: `"id"=1`,
			},
		*/

	}
}

type filterTestCase struct {
	filter interface{}
	result string
	err    bool
}

func (f *filterTestCase) Fail(output string, err error) string {
	return fmt.Sprintf("Error getting filter for %v expected=%v actual=%v shoulderr=%v err=%v", f.filter, f.result, output, f.err, err)
}

func getTestStorage() (*Storage, error) {
	b, err := ioutil.ReadFile("../../test_metadata.json")
	if err != nil {
		return nil, err
	}
	metaFunc, err := metadata.StaticMetaFunc(string(b))
	if err != nil {
		return nil, err
	}

	return &Storage{
		metaFunc: metaFunc,
	}, nil
}

func getFilter(filter interface{}) query.QueryArgs {
	return query.QueryArgs{
		Filter:        filter,
		DB:            "example_forum",
		ShardInstance: "dbshard_example_forum_2",
		Collection:    "user",
	}
}

func TestFilterToWhere(t *testing.T) {
	s, err := getTestStorage()
	if err != nil {
		t.Fatalf("Error getting test storage: %v", err)
	}

	for _, filterTest := range filterTestCases {
		ret, err := s.filterToWhere(getFilter(filterTest.filter))
		if err != nil {
			if !filterTest.err {
				t.Errorf(filterTest.Fail(ret, err))
			}
		} else {
			if filterTest.err {
				t.Errorf(filterTest.Fail(ret, err))
			}
			if ret != filterTest.result {
				t.Errorf(filterTest.Fail(ret, err))
			}
		}
	}

}
