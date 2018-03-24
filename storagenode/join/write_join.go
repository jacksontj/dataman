package join

import (
	"context"
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/client"
	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/query"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

// Struct to encapsulate returns of writes from subrecords done
type WriteJoinSubrecord struct {
	Path  []string
	Value interface{}
}

func (w *WriteJoinSubrecord) Apply(record record.Record) {
	record.Set(w.Path, w.Value)
}

// TODO: change this to have a flag on whether to stop on error or continue (or maybe switch to channels for the error? then we can decide either way
// return map[path]subrecord validationError, error
func DoWriteJoins(ctx context.Context, client *datamanclient.Client, q *query.Query, meta *metadata.Meta, collection *metadata.Collection, joinField interface{}, record record.Record) ([]*WriteJoinSubrecord, error, error) {
	joinMap, err := ParseJoinMap(joinField)
	if err != nil {
		return nil, nil, err
	}

	getter := func(name string) (MetaCollection, error) {
		return meta.GetCollection(q.Args.DB, q.Args.ShardInstance, name)
	}

	joinCollection, err := OrderJoins(getter, collection, joinMap)
	if err != nil {
		return nil, nil, err
	}

	return DoWriteJoin(ctx, client, q, joinCollection, record)
}

func DoWriteJoin(ctx context.Context, client *datamanclient.Client, q *query.Query, joinCollection *Collection, record record.Record) ([]*WriteJoinSubrecord, error, error) {
	writeJoinRecords := make([]*WriteJoinSubrecord, 0)

	// Do forward joins
	for _, forwardJoin := range joinCollection.ForwardJoin {
		joinKeyParts := strings.Split(forwardJoin.Key, ".")
		joinKeyParts[len(joinKeyParts)-1] += "."

		if rawRecord, _ := record.Pop(joinKeyParts); rawRecord != nil {
			// TODO: check that we can assert
			subRecords := rawRecord.([]interface{})
			// Go depth first so that things get removed
			replacementSubRecords := make([]interface{}, len(subRecords))
			for i, rawSubRecord := range subRecords {
				subRecord := rawSubRecord.(map[string]interface{})

				var subRecordWrites []*WriteJoinSubrecord

				if forwardJoin.C.HasJoins() {
					swrites, validationErr, err := DoWriteJoin(ctx, client, q, forwardJoin.C, subRecord)
					if validationErr != nil || err != nil {
						return nil, validationErr, err
					}
					subRecordWrites = swrites
				}

				if len(subRecord) == 1 {
					if subRecordWrites != nil {
						for _, subRecordWrite := range subRecordWrites {
							subRecordWrite.Apply(subRecord)
						}
					}
					replacementSubRecords[i] = subRecord
					continue
				}

				// At this point we need to do our layer
				joinResults, err := client.DoQuery(ctx, &query.Query{
					Type: query.Set,
					Args: query.QueryArgs{
						DB:            q.Args.DB,
						ShardInstance: q.Args.ShardInstance,
						Collection:    forwardJoin.C.Name,
						Record:        subRecord,
					},
				})
				if err != nil {
					// TODO: better -- right now we are only using local which can't have a transport error, but this
					// is a library method so we need to deal with it
					return nil, nil, err
				}

				if joinResults.ValidationError != nil {
					return nil, fmt.Errorf("%v", joinResults.ValidationError), joinResults.Err()
				}

				if err := joinResults.Err(); err != nil {
					return nil, nil, err
				}

				if subRecordWrites != nil {
					for _, subRecordWrite := range subRecordWrites {
						subRecordWrite.Apply(joinResults.Return[0])
					}
				}

				replacementSubRecords[i] = joinResults.Return[0]
			}

			writeJoinRecords = append(writeJoinRecords, &WriteJoinSubrecord{
				Path:  joinKeyParts,
				Value: replacementSubRecords,
			})

		} else {
			return nil, nil, fmt.Errorf("WriteJoin unable to find key %s in %v", forwardJoin.Key, record)
		}

	}

	// Do reverse joins
	for _, reverseJoin := range joinCollection.ReverseJoin {
		recordKey := reverseJoin.C.Name + "." + reverseJoin.JoinField.FullName()
		if rawRecord, _ := record.Pop([]string{recordKey}); rawRecord != nil {

			// rawRecord is the value for the join we are about to do.

			// TODO: check that we can assert
			subRecords := rawRecord.([]interface{})

			replacementSubRecords := make([]interface{}, len(subRecords))
			for i, rawSubRecord := range subRecords {
				subRecord := rawSubRecord.(map[string]interface{})

				var subRecordWrites []*WriteJoinSubrecord

				if reverseJoin.C.HasJoins() {
					swrites, validationErr, err := DoWriteJoin(ctx, client, q, reverseJoin.C, subRecord)
					if validationErr != nil || err != nil {
						return nil, validationErr, err
					}
					subRecordWrites = swrites
				}

				if len(subRecord) == 1 {
					if subRecordWrites != nil {
						for _, subRecordWrite := range subRecordWrites {
							subRecordWrite.Apply(subRecord)
						}
					}
					replacementSubRecords[i] = subRecord
					continue
				}

				// At this point we need to do our layer
				joinResults, err := client.DoQuery(ctx, &query.Query{
					Type: query.Set,
					Args: query.QueryArgs{
						DB:            q.Args.DB,
						ShardInstance: q.Args.ShardInstance,
						Collection:    reverseJoin.C.Name,
						Record:        subRecord,
					},
				})
				if err != nil {
					// TODO: better -- right now we are only using local which can't have a transport error, but this
					// is a library method so we need to deal with it
					return nil, nil, err
				}

				if joinResults.ValidationError != nil {
					return nil, fmt.Errorf("%v", joinResults.ValidationError), joinResults.Err()
				}

				if err := joinResults.Err(); err != nil {
					return nil, nil, err
				}

				if subRecordWrites != nil {
					for _, subRecordWrite := range subRecordWrites {
						subRecordWrite.Apply(joinResults.Return[0])
					}
				}

				replacementSubRecords[i] = joinResults.Return[0]
			}

			writeJoinRecords = append(writeJoinRecords, &WriteJoinSubrecord{
				Path:  []string{recordKey},
				Value: replacementSubRecords,
			})
		} else {
			return nil, nil, fmt.Errorf("WriteJoin unable to find reverse-join key %s in %v", reverseJoin.Key, record)
		}

	}

	return writeJoinRecords, nil, nil
}
