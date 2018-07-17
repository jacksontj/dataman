package record

import (
	"time"

	"github.com/jacksontj/dataman/datamantype"
)

type RecordItem struct {
	Record Record
	Source int
}

func NewRecordHeap(splitSortKeys [][]string, reverseList []bool) *RecordHeap {
	return &RecordHeap{
		Heap:          make([]RecordItem, 0),
		splitSortKeys: splitSortKeys,
		reverseList:   reverseList,
	}
}

// RecordHeap is a heap for use in sorting Record objects
// Records need "special" sorting as we potentially have many
// fields to sort by, so we effectively need to sort by each key
// until one has a comparison that sorts -- otherwise we continue down
// the list of sortKeys until we find one or we hit the end.
// In addition to the sort we also need to support "reverse" but we need the
// heap to still work with pop/push. To make this work we just have a reverseList
// option per sortKey, and the underlying sort list will negate the Less() return
// if "reverse" is true
type RecordHeap struct {
	Heap          []RecordItem
	splitSortKeys [][]string
	reverseList   []bool
}

func (r RecordHeap) Len() int { return len(r.Heap) }

func (r RecordHeap) Less(i, j int) (l bool) {
	var end bool
	var reverse bool
	// Generally we like to avoid named return variables, but in this case it
	// dramatically simplifies the code (since we need to support reversing).
	// We specifically want to reverse the result if we found a less than
	// value for a field in the sortKeys. If all of them matched and we got
	// to the "end" then we don't want to reverse the sort (as that would cause
	// unnecessary moving)
	defer func() {
		if reverse && !end {
			l = !l
		}
	}()
	for sortKeyIdx, keyParts := range r.splitSortKeys {
		reverse = r.reverseList[sortKeyIdx]
		// TODO: record could (and should) point at the CollectionFields which will tell us types
		iVal, _ := r.Heap[i].Record.Get(keyParts)
		jVal, _ := r.Heap[j].Record.Get(keyParts)
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
		case time.Time:
			jValTyped := jVal.(time.Time)
			if iValTyped != jValTyped {
				l = iValTyped.Before(jValTyped)
				return
			}
		case datamantype.Time:
			jValTyped := jVal.(datamantype.Time)
			if iValTyped != jValTyped {
				l = time.Time(iValTyped).Before(time.Time(jValTyped))
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
	end = true
	return
}

func (r RecordHeap) Swap(i, j int) { r.Heap[i], r.Heap[j] = r.Heap[j], r.Heap[i] }

func (r *RecordHeap) Push(x interface{}) {
	r.Heap = append(r.Heap, x.(RecordItem))
}

func (r *RecordHeap) Pop() interface{} {
	old := r.Heap
	n := len(old)
	x := old[n-1]
	r.Heap = old[0 : n-1]
	return x
}
