package metrics

import (
	"context"
	"fmt"
	"sync"

	radix "github.com/armon/go-radix"
)

func NewNamespaceRegistry(n string) *NamespaceRegistry {
	return &NamespaceRegistry{
		Namespace:  n,
		prefixTree: radix.New(),
		m:          &sync.Map{},
	}
}

type NamespaceRegistry struct {
	Namespace string

	l          sync.Mutex
	prefixTree *radix.Tree

	// Map of name -> collectable
	m *sync.Map
}

// Collect simply calls collect on all the collectables in this registry adding its
// namespace as a prefix to the name
func (n *NamespaceRegistry) Collect(ctx context.Context, points chan MetricPoint) error {
	f := func(_ string, c Collectable) error {
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
					item.Metric.Name = n.Namespace + "." + item.Metric.Name
				}
				points <- item
			}
		}
		return err
	}

	return n.Each(ctx, f)
}

func (n *NamespaceRegistry) Register(name string, c Collectable) error {
	n.l.Lock()
	defer n.l.Unlock()

	// If c is a registry, we need to add it to the prefix tree
	if _, ok := c.(Registry); ok {
		name += "."
	}

	if prefix, item, ok := n.prefixTree.LongestPrefix(name); ok {
		return fmt.Errorf("cannot register metric as it conflicts with a sub-namespace: %v %v", prefix, item)
	}

	if _, ok := n.m.LoadOrStore(name, c); !ok {
		// If c is a registry, we need to add it to the prefix tree
		if _, ok := c.(Registry); ok {
			n.prefixTree.Insert(name, c)
		}

		return nil
	} else {
		return fmt.Errorf("Collectable with that name already registered")
	}
}

func (n *NamespaceRegistry) Unregister(name string) error {
	n.m.Delete(name)
	return nil
}

func (n *NamespaceRegistry) Get(name string) Collectable {
	c, _ := n.m.Load(name)
	if c == nil {
		return nil
	}
	return c.(Collectable)
}

func (n *NamespaceRegistry) Each(ctx context.Context, eachFunc RegistryEachFunc) error {
	var err error

	f := func(kRaw, vRaw interface{}) bool {
		k := kRaw.(string)
		v := vRaw.(Collectable)

		err = eachFunc(k, v)
		if err != nil {
			return false
		}
		return true
	}
	n.m.Range(f)

	return err
}
