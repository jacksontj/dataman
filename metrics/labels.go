package metrics

import (
	"fmt"
	"sort"
	"strings"
)

type LabelSet map[string]string

func (l LabelSet) String() string {
	lstrs := make([]string, 0, len(l))
	for l, v := range l {
		lstrs = append(lstrs, fmt.Sprintf("%s=%q", l, v))
	}

	sort.Strings(lstrs)
	return fmt.Sprintf("{%s}", strings.Join(lstrs, ", "))

}

func MergeLabels(m LabelSet, k []string, v []string) LabelSet {
	n := make(LabelSet)

	for k, v := range m {
		n[k] = v
	}

	for i, kv := range k {
		n[kv] = v[i]
	}
	return n
}

func MergeLabelsDirect(m LabelSet, o LabelSet) LabelSet {
	n := make(LabelSet)

	for k, v := range m {
		n[k] = v
	}

	for k, v := range o {
		n[k] = v
	}
	return n
}
