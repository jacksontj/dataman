package metrics

import "strings"

// A metric is defined as (1) name and (2) labelset
type Metric struct {
	Name   string // TODO: remove from here?
	Labels LabelSet
	Help   string
}

// TODO: nicely layout the m.Labels (instead of the go print out)
func (m Metric) String() string {
	if len(m.Labels) == 0 {
		return m.Name
	}

	var sb strings.Builder
	sb.WriteString(m.Name)
	sb.WriteRune('{')

	i := len(m.Labels)
	for k, v := range m.Labels {
		sb.WriteString(k + `="` + v + `"`)
		i--
		if i > 0 {
			sb.WriteRune(',')
		}
	}
	sb.WriteRune('}')
	return sb.String()
}
