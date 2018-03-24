package query

import (
	"fmt"
	"strings"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/routernode/sharding"
)

// Encapsulate a result from the datastore
type Result struct {
	Return []record.Record `json:"return"`
	Errors []string        `json:"errors,omitempty"`
	// TODO: pointer to the right thing
	ValidationError interface{}            `json:"validation_error,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
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
// This function is responsible for maintaining sort order (if passed in args)
func MergeResult(pkeyFields []string, numResults int, results chan *Result) *Result {
	// Fast-path single results
	if numResults == 1 {
		r := <-results
		return r
	}

	pkeyFieldParts := make([][]string, len(pkeyFields))
	for i, pkeyField := range pkeyFields {
		pkeyFieldParts[i] = strings.Split(pkeyField, ".")
	}

	// We want to make sure we don't duplicate return entries
	ids := make(map[uint64]struct{})

	combinedResult := &Result{
		Return: make([]record.Record, 0),
		Meta:   make(map[string]interface{}),
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
