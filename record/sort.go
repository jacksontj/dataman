package record

import (
	"sort"
	"strings"
)

// sort the given data by the given keys
func Sort(sortKeys []string, reverseList []bool, data []Record) {
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
			iVal, _ := data[i].Get(keyParts)
			jVal, _ := data[j].Get(keyParts)
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
