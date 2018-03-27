package metrics

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sync"
)

func NewValuerArray(m Metric, c ValuerCreator, l []string) *ValuerArray {
	return &ValuerArray{
		Metric:    m,
		Creator:   c,
		LabelKeys: l,
	}
}

// Store an array of metrics.
type ValuerArray struct {
	// Base name + labelset to apply to all sub-metrics
	Metric

	// Function to create a new Value
	Creator ValuerCreator

	// Keys of all the labels allowed for metrics in this array
	LabelKeys []string

	// Map of labelset-hash -> Valuer
	//m map[uint64]Valuer
	m sync.Map
	// Map of labelset-hash -> label values
	// uint64->[]string
	mL sync.Map
}

func (m *ValuerArray) Name() string {
	return m.Metric.Name
}

func (m *ValuerArray) Collect(ctx context.Context, c chan<- MetricPoint) error {
	var err error
	f := func(kRaw, vRaw interface{}) bool {
		k := kRaw.(uint64)
		v := vRaw.(Valuer)

		labelValues, ok := m.mL.Load(k)
		if !ok {
			err = fmt.Errorf("Unable to get label values")
			return false
		}

		c <- MetricPoint{
			Metric{
				Name:   m.Metric.Name,
				Labels: MergeLabels(m.Metric.Labels, m.LabelKeys, labelValues.([]string)),
			},
			v.Value(),
		}
		return true
	}

	m.m.Range(f)
	return nil
}

// Access it by the slice of values
func (m *ValuerArray) WithValues(vals ...string) Valuer {
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

	valuer, ok := m.m.Load(s)
	if ok {
		return valuer.(Valuer)
	} else {
		// Otherwise it doesn't exist, so lets try adding it
		if _, ok := m.mL.LoadOrStore(s, vals); ok {
			return m.WithValues(vals...)
		}
		valuer = m.Creator()
		if _, ok = m.m.LoadOrStore(s, valuer); ok {
			return m.WithValues(vals...)
		}
		return valuer.(Valuer)
	}
}

func (m *ValuerArray) CounterWithValues(vals ...string) CounterType {
	v := m.WithValues(vals...)
	return v.(CounterType)
}

func (m *ValuerArray) GaugeWithValues(vals ...string) GaugeType {
	v := m.WithValues(vals...)
	return v.(GaugeType)
}

func (m *ValuerArray) ObserveWithValues(vals ...string) ObserveType {
	v := m.WithValues(vals...)
	return v.(ObserveType)
}
