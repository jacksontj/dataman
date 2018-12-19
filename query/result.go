package query

import (
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/routernode/sharding"
	"github.com/jacksontj/dataman/storagenode/metadata/aggregation"
)

// Encapsulate a result from the datastore
type Result struct {
	Return []record.Record `json:"return"`
	Errors []string        `json:"errors,omitempty"`
	// TODO: pointer to the right thing
	ValidationError interface{}       `json:"validation_error,omitempty"`
	Meta            map[string]string `json:"meta,omitempty"`
}

func (r *Result) Err() error {
	if r.Errors == nil {
		return nil
	} else {
		return fmt.Errorf(strings.Join(r.Errors, "\n"))
	}
}

func (r *Result) Sort(keys []string, reverseList []bool) {
	if r.Return != nil {
		record.Sort(keys, reverseList, r.Return)
	}
}

func (r *Result) Project(fields []string) {
	projectionFields := record.ProjectionFields(fields)

	for i, returnRecord := range r.Return {
		r.Return[i] = returnRecord.Project(projectionFields)
	}
}

// Merge multiple results together
func MergeResult(primaryIndexFields []string, numResults int, results chan *Result) *Result {
	// Fast-path single results
	if numResults == 1 {
		r := <-results
		return r
	}

	pkeyFieldParts := make([][]string, len(primaryIndexFields))
	for i, pkeyField := range primaryIndexFields {
		pkeyFieldParts[i] = strings.Split(pkeyField, ".")
	}

	// We want to make sure we don't duplicate return entries
	ids := make(map[uint64]struct{})

	combinedResult := &Result{
		Return: make([]record.Record, 0),
		Meta:   make(map[string]string),
	}

	recievedResults := 0
	for result := range results {
		if result.Errors != nil {
			if combinedResult.Errors == nil {
				combinedResult.Errors = result.Errors
			} else {
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}
		}
		// TODO: merge meta
		if len(combinedResult.Meta) == 0 {
			combinedResult.Meta = result.Meta
		}

		for _, resultReturn := range result.Return {

			pkeyFields := make([]interface{}, len(pkeyFieldParts))
			var ok bool
			for i, pkeyField := range pkeyFieldParts {
				pkeyFields[i], ok = resultReturn.Get(pkeyField)
				if !ok {
					// TODO: something else?
					panic("Missing pkey in response!!!")
				}
			}
			pkey, err := (sharding.HashMethod(sharding.SHA256).Get())(sharding.CombineKeys(pkeyFields))
			if err != nil {
				panic(fmt.Sprintf("MergeResult doesn't know how to hash pkey: %v", err))
			}
			if _, ok := ids[pkey]; !ok {
				ids[pkey] = struct{}{}
				combinedResult.Return = append(combinedResult.Return, resultReturn)
			}
		}
		recievedResults++
		if recievedResults == numResults {
			break
		}
	}

	return combinedResult
}

// Merge multiple aggregation results together
func MergeAggregateResult(args QueryArgs, numResults int, results chan *Result) *Result {
	// Fast-path single results
	if numResults == 1 {
		r := <-results
		return r
	}

	uniqFieldParts := make([][]string, 0)
	for fieldPath, aggregationList := range args.AggregationFields {
		if len(aggregationList) == 0 {
			uniqFieldParts = append(uniqFieldParts, strings.Split(fieldPath, "."))
		}
	}

	// mapping of uniqField -> record
	ids := make(map[uint64]record.Record)

	combinedResult := &Result{
		Return: make([]record.Record, 0),
		Meta:   make(map[string]string),
	}

	recievedResults := 0
	for result := range results {
		if result.Errors != nil {
			if combinedResult.Errors == nil {
				combinedResult.Errors = result.Errors
			} else {
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}
		}
		// TODO: merge meta
		if len(combinedResult.Meta) == 0 {
			combinedResult.Meta = result.Meta
		}

		for _, resultReturn := range result.Return {

			var ukey uint64
			if len(uniqFieldParts) > 0 {
				var ok bool
				uniqFields := make([]interface{}, len(uniqFieldParts))
				for i, uniqField := range uniqFieldParts {
					uniqFields[i], ok = resultReturn.Get(uniqField)
					if !ok {
						// TODO: something else?
						panic("Missing uniqField in response!!!")
					}
				}
				var err error
				ukey, err = (sharding.HashMethod(sharding.SHA256).Get())(sharding.CombineKeys(uniqFields))
				if err != nil {
					panic(fmt.Sprintf("MergeResult doesn't know how to hash pkey: %v", err))
				}
			}
			// if we haven't seen one yet, just save it
			if existingRecord, ok := ids[ukey]; !ok {
				ids[ukey] = resultReturn
				combinedResult.Return = append(combinedResult.Return, resultReturn)
			} else {
				for fieldName, aggregationList := range args.AggregationFields {
					if aggregationList == nil {
						continue
					}
					fieldParts := strings.Split(fieldName, ".")
					for _, aggregationMethod := range aggregationList {
						origLastFieldPart := fieldParts[len(fieldParts)-1]
						fieldParts[len(fieldParts)-1] += "." + string(aggregationMethod)

						// get the existing and new Values
						existingValue, ok := existingRecord.Get(fieldParts)
						if !ok {
							panic("missing agg field")
						}
						newVal, ok := resultReturn.Get(fieldParts)
						if !ok {
							panic("missing agg field")
						}

						switch aggregationMethod {
						case aggregation.Count:
							// TODO: check return
							// TODO: know what the types are?
							existingRecord.Set(fieldParts, existingValue.(float64)+newVal.(float64))

						case aggregation.Sum:
							// TODO: check return
							// TODO: know what the types are?
							existingRecord.Set(fieldParts, existingValue.(float64)+newVal.(float64))

						case aggregation.Min:
							// TODO: check return
							// TODO: know what the types are?
							if newVal.(float64) < existingValue.(float64) {
								existingRecord.Set(fieldParts, newVal.(float64))
							}

						case aggregation.Max:
							// TODO: check return
							// TODO: know what the types are?
							if newVal.(float64) > existingValue.(float64) {
								existingRecord.Set(fieldParts, newVal.(float64))
							}

						default:
							panic("Unknown aggregation method")
						}
						fieldParts[len(fieldParts)-1] = origLastFieldPart
					}

				}
			}
		}

		recievedResults++
		if recievedResults == numResults {
			break
		}
	}

	return combinedResult
}
