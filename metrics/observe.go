package metrics

import "fmt"

// type specific array collectable
func NewCustomObserveArray(m Metric, c CollectableCreator, l []string) (*ObserveArray, error) {
	if _, ok := c().(ObserveType); !ok {
		return nil, fmt.Errorf("CollectableCreator must generate an item which is an ObserveType")
	}

	return &ObserveArray{NewCollectableArray(m, c, l)}, nil
}

type ObserveArray struct {
	*CollectableArray
}

func (g *ObserveArray) WithValues(vals ...string) ObserveType {
	r := g.CollectableArray.WithValues(vals...)
	return r.(ObserveType)
}
