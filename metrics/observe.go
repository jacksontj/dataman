package metrics

// type specific array collectable
func NewObserveArray(m Metric, c CollectableCreator, l []string) *ObserveArray {
	if _, ok := c().(ObserveType); !ok {
		panic("c must return GaugeType")
	}

	return &ObserveArray{NewCollectableArray(m, c, l)}
}

type ObserveArray struct {
	*CollectableArray
}

func (g *ObserveArray) WithValues(vals ...string) ObserveType {
	r := g.CollectableArray.WithValues(vals...)
	return r.(ObserveType)
}
