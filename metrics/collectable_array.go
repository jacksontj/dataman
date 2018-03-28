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

	// Map of labelset-hash -> Valuer
	//m map[uint64]Valuer
	m sync.Map
	// Map of labelset-hash -> label values
	// uint64->[]string
	mL sync.Map
}

func (m *CollectableArray) Describe(c chan<- MetricDesc) error {
	var err error
	ch := make(chan MetricDesc)
	go func() {
		defer close(ch)
		err = m.Creator().Describe(ch)
	}()

	for d := range ch {
		if d.Name != "" {
			d.Name = m.Metric.Name + "_" + d.Name
		} else {
			d.Name = m.Metric.Name
		}
		c <- d
	}
	return err
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
			func(point *MetricPoint) error {
				name := m.Metric.Name
				if point.Metric.Name != "" {
					name += "_" + point.Metric.Name
				}
				*point = MetricPoint{
					Metric{
						Name: name,
						Labels: MergeLabelsDirect(
							MergeLabels(m.Metric.Labels, m.LabelKeys, labelValues.([]string)),
							point.Metric.Labels),
					},
					point.Value,
				}
				return nil
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

// Access it by the slice of values
func (m *CollectableArray) WithValues(vals ...string) Collectable {

	h := sha1.New()

	for _, v := range vals {
		h.Write([]byte(v))
		h.Write([]byte(","))
	}

	// TODO: use hashing from sharding package?
	sum := h.Sum(nil)
	var buf []byte
	buf = sum[:]
	s := binary.LittleEndian.Uint64(buf)

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
