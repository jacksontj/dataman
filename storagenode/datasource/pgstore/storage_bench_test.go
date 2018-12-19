package pgstorage

import (
	"strconv"
	"strings"
	"testing"

	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

func BenchmarkFilterToWhere(b *testing.B) {
	tests := []struct {
		meta metadata.Meta
		args []query.QueryArgs
	}{
		{
			meta: metadata.Meta{
				Databases: map[string]*metadata.Database{
					"test": &metadata.Database{
						ShardInstances: map[string]*metadata.ShardInstance{
							"1": &metadata.ShardInstance{
								Collections: map[string]*metadata.Collection{
									"user": &metadata.Collection{
										Fields: map[string]*metadata.CollectionField{
											"id": &metadata.CollectionField{
												Type: "_int",
											},
											"username": &metadata.CollectionField{
												Type: "_string",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: []query.QueryArgs{
				{
					DB:            "test",
					ShardInstance: "1",
					Collection:    "user",
					Filter:        map[string]interface{}{"id": []interface{}{"=", 100}},
				},
				{
					DB:            "test",
					ShardInstance: "1",
					Collection:    "user",
					Filter:        map[string]interface{}{"id": []interface{}{"=", 100}, "username": []interface{}{"=", "testuser"}},
				},
				{
					DB:            "test",
					ShardInstance: "1",
					Collection:    "user",
					Filter: []interface{}{
						map[string]interface{}{"id": []interface{}{"=", 100}, "username": []interface{}{"=", "testuser"}},
						"OR",
						map[string]interface{}{"id": []interface{}{"=", 100}, "username": []interface{}{"=", "testuser"}},
					},
				},
			},
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			s := &Storage{
				metaFunc: func() *metadata.Meta { return &test.meta },
			}
			for j, arg := range test.args {
				b.Run(strconv.Itoa(j), func(b *testing.B) {
					for x := 0; x < b.N; x++ {
						s.filterToWhere(arg)
					}
				})
			}
			for j, arg := range test.args {
				b.Run(strconv.Itoa(j)+"_builder", func(b *testing.B) {
					for x := 0; x < b.N; x++ {
						b := strings.Builder{}
						s.filterToWhereBuilder(&b, arg)
					}
				})
			}
		})
	}
}
