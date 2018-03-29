package storagenode

import (
	"github.com/jacksontj/dataman/metrics"
)

func NewStorageNodeMetrics(r metrics.Registry) StorageNodeMetrics {
	m := StorageNodeMetrics{}

	return m
}

type StorageNodeMetrics struct {
}
