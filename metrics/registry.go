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
	n.l.Lock()
	defer n.l.Unlock()

	switch cTyped := c.(type) {
	case NamedCollectable:
		name := cTyped.Name()

		if name == "" {
			return fmt.Errorf("NamedCollectable must have a name")
		}
		if prefix, item, ok := n.prefixTree.LongestPrefix(name); ok {
			return fmt.Errorf("cannot register metric as it conflicts with a sub-namespace: %v %v", prefix, item)
		}

		if _, ok := n.m.LoadOrStore(name, c); !ok {
			return nil
		} else {
			return fmt.Errorf("Collectable with that name already registered")
		}

	case PrefixCollectable:
		prefix := cTyped.Prefix() + "_"

		if prefix == "" {
			return fmt.Errorf("PrefixCollectable must have a prefix")
		}
		if prefix, item, ok := n.prefixTree.LongestPrefix(prefix); ok {
			return fmt.Errorf("cannot register metric as it conflicts with a sub-namespace: %v %v", prefix, item)
		}

		if _, ok := n.m.LoadOrStore(prefix, c); !ok {
			n.prefixTree.Insert(prefix, c)
			return nil
		} else {
			return fmt.Errorf("Collectable with that name already registered")
		}
	default:
		return fmt.Errorf("Unsupported collectable")
	}

}

func (n *NamespaceRegistry) Prefix() string {
	return n.Namespace
}

func (n *NamespaceRegistry) Unregister(c Collectable) error {
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
