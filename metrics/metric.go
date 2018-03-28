package metrics

import (
	"fmt"
)

// Description of metrics
type MetricDesc struct {
	Name   string
	Prefix bool
}

// A metric is defined as (1) name and (2) labelset
type Metric struct {
	Name   string // TODO: remove from here?
	Labels LabelSet
	Help   string
}

// TODO: nicely layout the m.Labels (instead of the go print out)
func (m Metric) String() string {
	return fmt.Sprintf("%s%v", m.Name, m.Labels)
}
