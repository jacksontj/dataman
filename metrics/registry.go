package metrics

import (
	"context"
	"sync"
)

func NewNamespaceRegistry(n string) *NamespaceRegistry {
	return &NamespaceRegistry{
		Namespace: n,
		mr:        NewMetricDescRegistry(),
		m:         &sync.Map{},
	}
}

type NamespaceRegistry struct {
	Namespace string

	mr *MetricDescRegistry

	// TODO: have a name??
	// Map of collectable -> collectable
	m *sync.Map
}

func (n *NamespaceRegistry) Describe(c chan<- MetricDesc) error {
	c <- MetricDesc{
		Name:   n.Namespace,
		Prefix: true,
	}
	return nil
}

// Collect simply calls collect on all the collectables in this registry adding its
// namespace as a prefix to the name
func (n *NamespaceRegistry) Collect(ctx context.Context, points chan<- MetricPoint) error {
	f := func(c Collectable) error {
		var err error
		// We need to call collect on the children and add our namespace stuff to the value that is returned
		innerPoints := make(chan MetricPoint)
		go func() {
			defer close(innerPoints)
			err = c.Collect(ctx, innerPoints)
		}()

	WAITRESULT:
		for {
			select {
			case item, ok := <-innerPoints:
				if !ok {
					break WAITRESULT
				}
				if n.Namespace != "" {
					item.Metric.Name = n.Namespace + "_" + item.Metric.Name
				}
				points <- item
			}
		}
		return err
	}

	return n.Each(ctx, f)
}

func (n *NamespaceRegistry) Register(c Collectable) error {
	var err error
	ch := make(chan MetricDesc)
	go func() {
		defer close(ch)
		err = c.Describe(ch)
	}()

	metricDescs := make([]MetricDesc, 0)
	for metricDesc := range ch {
		metricDescs = append(metricDescs, metricDesc)
	}

	if err != nil {
		return err
	}

	// check that we can register all the names
	if err := n.mr.AddOrError(metricDescs); err != nil {
		return err
	}

	if _, ok := n.m.LoadOrStore(c, c); ok {
		panic("shouldn't be possible!")
	}

	return nil
}

/*
func (n *NamespaceRegistry) Unregister(c Collectable) error {
	var err error
	ch := make(chan MetricDesc)
	go func() {
		defer close(ch)
		err = c.Describe(ch)
	}()

	metricDescs := make([]MetricDesc, 0)
	for metricDesc := range ch {
		metricDescs = append(metricDescs, metricDesc)
	}

	if err != nil {
		return err
	}

	switch cTyped := c.(type) {
	case NamedCollectable:
		n.m.Delete(cTyped.Name())

	case PrefixCollectable:
		n.l.Lock()
		defer n.l.Unlock()

		prefix := cTyped.Prefix() + "_"

		n.m.Delete(prefix)
		n.prefixTree.Delete(prefix)
	default:
		return fmt.Errorf("Unsupported collectable")
	}

	return nil
}
*/

func (n *NamespaceRegistry) Each(ctx context.Context, eachFunc RegistryEachFunc) error {
	var err error

	f := func(_, vRaw interface{}) bool {
		v := vRaw.(Collectable)

		err = eachFunc(v)
		if err != nil {
			return false
		}
		return true
	}
	n.m.Range(f)

	return err
}
