package metrics

import (
	"fmt"
	"sync"

	radix "github.com/armon/go-radix"
)

// TODO: better name, we don't want to call this a registry since this doesn't
// implement the Registry interface
func NewMetricDescRegistry() *MetricDescRegistry {
	return &MetricDescRegistry{
		prefixTree: radix.New(),
	}
}

// TODO: make these mergeable
type MetricDescRegistry struct {
	l sync.RWMutex
	// Tree of prefix -> bool (prefix or not)
	prefixTree *radix.Tree
}

func (n *MetricDescRegistry) Contains(name string) bool {
	n.l.RLock()
	defer n.l.RUnlock()

	prefix, item, ok := n.prefixTree.LongestPrefix(name)
	// If we don't have any prefix matching, then we don't have this
	if !ok {
		return false
	}

	// If the prefix matches, regardless its contained
	if prefix == name {
		return true
	}

	// If its not an exact match, but the matching one is a prefix -- its a match
	return item.(MetricDesc).Prefix
}

func (n *MetricDescRegistry) AddOrError(ds []MetricDesc) error {
	n.l.Lock()
	defer n.l.Unlock()
	// check if we can add these
	for _, d := range ds {
		if !n.canAddDesc(d) {
			// TODO: nicer error about which metric?
			return fmt.Errorf("Unable to add %v, namespace already taken", d)
		}
	}

	// Assuming we can, lets do so!
	for _, d := range ds {
		// TODO: make sure we didn't update?
		n.prefixTree.Insert(d.Name, d)
	}

	return nil
}

// TODO: move to separate datastructure
func (n *MetricDescRegistry) canAddDesc(d MetricDesc) bool {
	prefix, item, ok := n.prefixTree.LongestPrefix(d.Name)
	// If we have nothing like this as a prefix
	if !ok {
		return true
	}

	// If they match, then we can't have it regardless of prefix
	if prefix == d.Name {
		return false
	}

	// If the matching one is a prefix, then we can't (since we'd collide)
	if item.(MetricDesc).Prefix {
		return false
	}

	// Otherwise we are all set
	return true
}

func (n *MetricDescRegistry) Remove(ds []MetricDesc) {
	n.l.Lock()
	defer n.l.Unlock()

	for _, d := range ds {
		n.prefixTree.Delete(d.Name)
	}
}

func (n *MetricDescRegistry) List() []MetricDesc {
	r := make([]MetricDesc, 0, n.prefixTree.Len())

	for _, v := range n.prefixTree.ToMap() {
		r = append(r, v.(MetricDesc))
	}
	return r
}
