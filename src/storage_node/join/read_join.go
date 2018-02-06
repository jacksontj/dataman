package join

import (
	"context"
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/src/client"
	"github.com/jacksontj/dataman/src/query"
	"github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/jacksontj/dataman/src/storage_node/metadata/filter"
)

// TODO: change this to have a flag on whether to stop on error or continue (or maybe switch to channels for the error? then we can decide either way
func DoReadJoins(ctx context.Context, client *datamanclient.Client, q *query.Query, meta *metadata.Meta, collection *metadata.Collection, joinField interface{}, records []map[string]interface{}) error {
	joinMap, err := ParseJoinMap(joinField)
	if err != nil {
		return err
	}

	getter := func(name string) (MetaCollection, error) {
		return meta.GetCollection(q.Args["db"].(string), q.Args["shard_instance"].(string), name)
	}

	joinCollection, err := OrderJoins(getter, collection, joinMap)
	if err != nil {
		return err
	}

	for i, record := range records {
		if err := DoReadJoin(ctx, client, q, joinCollection, record); err != nil {
			return err
		}
		records[i] = record
	}
	return nil
}

// TODO: reimplement to do records at once -- so we can batch?
func DoReadJoin(ctx context.Context, client *datamanclient.Client, q *query.Query, joinCollection *Collection, record map[string]interface{}) error {
	// Do forward join
	for _, forwardJoin := range joinCollection.ForwardJoin {
		joinKeyParts := strings.Split(forwardJoin.Key, ".")
		if rawRecord, _ := query.GetValue(record, joinKeyParts); rawRecord != nil {
			// add joinkey to the filter defined
			forwardJoin.Filter[forwardJoin.JoinField.Relation.Field] = []interface{}{filter.Equal, rawRecord}
			joinResults, err := client.DoQuery(ctx, &query.Query{
				Type: query.Filter,
				Args: map[string]interface{}{
					"db":             q.Args["db"],
					"shard_instance": q.Args["shard_instance"].(string),
					"collection":     forwardJoin.C.Name,
					"filter":         forwardJoin.Filter,
				},
			})

			if err != nil {
				// TODO: better -- right now we are only using local which can't have a transport error, but this
				// is a library method so we need to deal with it
				return err
			}

			if err := joinResults.Err(); err != nil {
				return err
			}

			if forwardJoin.C.HasJoins() {
				for i, resultRecord := range joinResults.Return {
					if err := DoReadJoin(ctx, client, q, forwardJoin.C, resultRecord); err == nil {
						joinResults.Return[i] = resultRecord
					} else {
						return err
					}
				}
			}

			joinKeyParts[len(joinKeyParts)-1] += "."
			query.SetValue(record, joinResults.Return, joinKeyParts)

		} else {
			return fmt.Errorf("ReadJoin unable to find forward-join key %s in %v", forwardJoin.Key, record)
		}

	}

	// Do reverse join
	for _, reverseJoin := range joinCollection.ReverseJoin {
		// Name of field in this record to do the join with

		if rawRecord, _ := query.GetValue(record, strings.Split(reverseJoin.JoinField.Relation.Field, ".")); rawRecord != nil {
			reverseJoin.Filter[reverseJoin.Key] = []interface{}{filter.Equal, rawRecord}
			joinResults, err := client.DoQuery(ctx, &query.Query{
				Type: query.Filter,
				Args: map[string]interface{}{
					"db":             q.Args["db"],
					"shard_instance": q.Args["shard_instance"].(string),
					"collection":     reverseJoin.C.Name,
					"filter":         reverseJoin.Filter,
				},
			})
			if err != nil {
				// TODO: better -- right now we are only using local which can't have a transport error, but this
				// is a library method so we need to deal with it
				return err
			}
			if err := joinResults.Err(); err != nil {
				return err
			}

			// Check if the child has stuff to do
			if reverseJoin.C.HasJoins() {
				for i, resultRecord := range joinResults.Return {
					if err := DoReadJoin(ctx, client, q, reverseJoin.C, resultRecord); err == nil {
						joinResults.Return[i] = resultRecord
					} else {
						return err
					}
				}
			}

			record[reverseJoin.C.Name+"."+reverseJoin.Key] = joinResults.Return
		} else {
			return fmt.Errorf("ReadJoin unable to find reverse-join key %s in %v", reverseJoin.JoinField.Relation.Field, record)
		}

	}

	return nil
}
