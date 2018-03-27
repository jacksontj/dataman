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
