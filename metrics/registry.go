package metrics

import (
	"context"
	"fmt"
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

	// Map of collectable -> *MetricDescRegistry
	m *sync.Map
}

func (n *NamespaceRegistry) Describe(ctx context.Context, c chan<- MetricDesc) error {
	select {
	case c <- MetricDesc{
		Name:   n.Namespace,
		Prefix: true,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Collect simply calls collect on all the collectables in this registry adding its
// namespace as a prefix to the name
func (n *NamespaceRegistry) Collect(ctx context.Context, points chan<- MetricPoint) error {
	f := func(c Collectable, r *MetricDescRegistry) error {
		transformations := []MetricPointTransformation{
			func(point *MetricPoint) (bool, error) {
				if !r.Contains(point.Name) {
					fmt.Printf("Skipping item %s as it wasn't present for Describe() (not in %v): %v", point.Name, r.prefixTree.ToMap(), point)
					point = nil
					return false, nil // Don't stop, just print the error and skip the item
				}
				if n.Namespace != "" {
					point.Metric.Name = n.Namespace + "_" + point.Metric.Name
				}
				return true, nil
			},
		}

		return StreamMetricPoints(ctx, c, points, transformations)
	}

	return n.Each(ctx, f)
}

func (n *NamespaceRegistry) Register(c Collectable) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	ch := make(chan MetricDesc)
	go func() {
		defer close(ch)
		err = c.Describe(ctx, ch)
	}()

	metricDescs := make([]MetricDesc, 0)
	for metricDesc := range ch {
		metricDescs = append(metricDescs, metricDesc)
	}

	if err != nil {
		return err
	}

	r := NewMetricDescRegistry()
	r.AddOrError(metricDescs)

	// check that we can register all the names
	if err := n.mr.AddOrError(metricDescs); err != nil {
		return err
	}

	if _, ok := n.m.LoadOrStore(c, r); ok {
		panic("shouldn't be possible!")
	}

	return nil
}

func (n *NamespaceRegistry) Unregister(c Collectable) error {
	// TODO: lock?
	metricsDescRegisterRaw, ok := n.m.Load(c)
	if !ok {
		return fmt.Errorf("Unable to unregister as it wasn't registered")
	}

	metricsDescRegister := metricsDescRegisterRaw.(*MetricDescRegistry)

	descs := metricsDescRegister.List()

	// TODO: check for error?
	// remove entries from our register
	n.mr.Remove(descs)

	// Remove from the sync map
	n.m.Delete(c)

	return nil
}

func (n *NamespaceRegistry) Each(ctx context.Context, eachFunc RegistryEachFunc) error {
	var err error

	f := func(kRaw, vRaw interface{}) bool {
		k := kRaw.(Collectable)
		v := vRaw.(*MetricDescRegistry)

		err = eachFunc(k, v)
		if err != nil {
			return false
		}
		return true
	}
	n.m.Range(f)

	return err
}
