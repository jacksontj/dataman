package metrics

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sync"
)

func NewCollectableArray(m Metric, c CollectableCreator, l []string) *CollectableArray {
	for _, label := range l {
		if _, ok := m.Labels[label]; ok {
			panic("Cannot create ValuerArray with label in base and l")
		}
	}
	return &CollectableArray{
		Metric:    m,
		Creator:   c,
		LabelKeys: l,
	}
}

// Store an array of metrics.
type CollectableArray struct {
	// Base name + labelset to apply to all sub-metrics
	Metric

	// Function to create a new Value
	Creator CollectableCreator

	// Keys of all the labels allowed for metrics in this array
	LabelKeys []string

	// Map of labelset-hash -> Collectable
	//m map[uint64]Collectable
	m sync.Map
	// Map of labelset-hash -> label values
	// uint64->[]string
	mL sync.Map
}

func (m *CollectableArray) Describe(ctx context.Context, c chan<- MetricDesc) error {
	transformations := []MetricDescTransformation{
		func(d *MetricDesc) (bool, error) {
			if d.Name != "" {
				d.Name = m.Metric.Name + "_" + d.Name
			} else {
				d.Name = m.Metric.Name
			}
			// TODO: labels
			return true, nil
		},
	}
	return StreamMetricDescs(ctx, m.Creator(), c, transformations)
}

func (m *CollectableArray) Collect(ctx context.Context, c chan<- MetricPoint) error {
	var err error
	f := func(kRaw, vRaw interface{}) bool {
		k := kRaw.(uint64)
		v := vRaw.(Collectable)

		labelValues, ok := m.mL.Load(k)
		if !ok {
			err = fmt.Errorf("Unable to get label values")
			return false
		}

		transformations := []MetricPointTransformation{
			func(point *MetricPoint) (bool, error) {
				name := m.Metric.Name
				if point.Metric.Name != "" {
					name += "_" + point.Metric.Name
				}
				*point = MetricPoint{
					Metric: Metric{
						Name: name,
						Labels: MergeLabelsDirect(
							MergeLabels(m.Metric.Labels, m.LabelKeys, labelValues.([]string)),
							point.Metric.Labels),
					},
					Value: point.Value,
				}
				return true, nil
			},
		}

		err = StreamMetricPoints(ctx, v, c, transformations)
		if err != nil {
			return false
		}
		return true
	}

	m.m.Range(f)
	return nil
}

func (m *CollectableArray) hashLabelValues(vals []string) uint64 {
	h := sha1.New()

	for _, v := range vals {
		h.Write([]byte(v))
		h.Write([]byte(","))
	}

	// TODO: use hashing from sharding package?
	return binary.LittleEndian.Uint64(h.Sum(nil)[:])
}

func (m *CollectableArray) Remove(vals ...string) {
	if len(vals) != len(m.LabelKeys) {
		return
	}

	s := m.hashLabelValues(vals)

	m.m.Delete(s)
	m.mL.Delete(s)
}

// Access it by the slice of values
func (m *CollectableArray) WithValues(vals ...string) Collectable {
	if len(vals) != len(m.LabelKeys) {
		panic("number of label values must match LabelKeys")
	}

	s := m.hashLabelValues(vals)

	collectable, ok := m.m.Load(s)
	if ok {
		return collectable.(Collectable)
	} else {
		// Otherwise it doesn't exist, so lets try adding it
		if _, ok := m.mL.LoadOrStore(s, vals); ok {
			return m.WithValues(vals...)
		}
		collectable = m.Creator()
		if _, ok = m.m.LoadOrStore(s, collectable); ok {
			return m.WithValues(vals...)
		}
		return collectable.(Collectable)
	}
}

func (m *CollectableArray) CounterWithValues(vals ...string) CounterType {
	v := m.WithValues(vals...)
	return v.(CounterType)
}

func (m *CollectableArray) GaugeWithValues(vals ...string) GaugeType {
	v := m.WithValues(vals...)
	return v.(GaugeType)
}

func (m *CollectableArray) ObserveWithValues(vals ...string) ObserveType {
	v := m.WithValues(vals...)
	return v.(ObserveType)
}
