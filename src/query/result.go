package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jacksontj/dataman/src/router_node/sharding"
)

// Encapsulate a result from the datastore
type Result struct {
	Return []map[string]interface{} `json:"return"`
	Errors []string                 `json:"errors,omitempty"`
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
		Sort(keys, reverseList, r.Return)
	}
}

func (r *Result) Project(fields []string) {
	projectionFields := ProjectionFields(fields)

	for i, returnRow := range r.Return {
		r.Return[i] = Project(projectionFields, returnRow)
	}
}

// Merge multiple results together
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
		Return: make([]map[string]interface{}, 0),
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
				pkeyFields[i], ok = GetValue(resultReturn, pkeyField)
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

func GetValue(value map[string]interface{}, nameParts []string) (interface{}, bool) {
	val, ok := value[nameParts[0]]
	if !ok {
		return nil, ok
	}

	for _, namePart := range nameParts[1:] {
		typedVal, ok := val.(map[string]interface{})
		if !ok {
			return nil, ok
		}
		val, ok = typedVal[namePart]
		if !ok {
			return nil, ok
		}
	}
	return val, true
}

func SetValue(value map[string]interface{}, newValue interface{}, nameParts []string) interface{} {
	var val interface{}
	if len(nameParts) > 1 {
		val = value[nameParts[0]]
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			val = val.(map[string]interface{})[namePart]
		}

	} else {
		val = value
	}

	val.(map[string]interface{})[nameParts[len(nameParts)-1]] = newValue

	return val
}

func DelValue(value map[string]interface{}, nameParts []string) bool {
	var val interface{}
	if len(nameParts) > 1 {
		val = value[nameParts[0]]
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			val = val.(map[string]interface{})[namePart]
		}

	} else {
		val = value
	}

	_, ok := val.(map[string]interface{})[nameParts[len(nameParts)-1]]
	if ok {
		delete(val.(map[string]interface{}), nameParts[len(nameParts)-1])
	}
	return ok
}

func PopValue(value map[string]interface{}, nameParts []string) (interface{}, bool) {
	var val interface{}
	if len(nameParts) > 1 {
		val = value[nameParts[0]]
		for _, namePart := range nameParts[1 : len(nameParts)-1] {
			val = val.(map[string]interface{})[namePart]
		}

	} else {
		val = value
	}

	tmp, ok := val.(map[string]interface{})[nameParts[len(nameParts)-1]]
	if ok {
		delete(val.(map[string]interface{}), nameParts[len(nameParts)-1])
	}
	return tmp, ok
}

func FlattenResult(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch typedV := v.(type) {
		case map[string]interface{}:
			// get the submap as a flattened thing
			subMap := FlattenResult(typedV)
			for subK, subV := range subMap {
				result[k+"."+subK] = subV
			}
		default:
			result[k] = v
		}
	}
	return result
}

// TODO: change map[string]interface to a `Record` struct type
// sort the given data by the given keys
func Sort(sortKeys []string, reverseList []bool, data []map[string]interface{}) {
	splitSortKeys := make([][]string, len(sortKeys))
	for i, sortKey := range sortKeys {
		splitSortKeys[i] = strings.Split(sortKey, ".")
	}

	less := func(i, j int) (l bool) {
		var reverse bool
		defer func() {
			if reverse {
				l = !l
			}
		}()
		for sortKeyIdx, keyParts := range splitSortKeys {
			reverse = reverseList[sortKeyIdx]
			// TODO: record could (and should) point at the CollectionFields which will tell us types
			iVal, _ := GetValue(data[i], keyParts)
			jVal, _ := GetValue(data[j], keyParts)
			switch iValTyped := iVal.(type) {
			case string:
				jValTyped := jVal.(string)
				if iValTyped != jValTyped {
					l = iValTyped < jValTyped
					return
				}
			case int:
				jValTyped := jVal.(int)
				if iValTyped != jValTyped {
					l = iValTyped < jValTyped
					return
				}
			case int64:
				jValTyped := jVal.(int64)
				if iValTyped != jValTyped {
					l = iValTyped < jValTyped
					return
				}
			case float64:
				jValTyped := jVal.(float64)
				if iValTyped != jValTyped {
					l = iValTyped < jValTyped
					return
				}
			case bool:
				jValTyped := jVal.(bool)
				if iValTyped != jValTyped {
					l = !iValTyped && jValTyped
					return
				}
			// TODO: return error? At this point if all return false, I'm not sure what happens
			default:
				panic("Unknown type")
				l = false
				return

			}
		}
		l = false
		return
	}
	sort.Slice(data, less)
}
