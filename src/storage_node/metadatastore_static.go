package storagenode

import (
	"context"
	"sync"

	"github.com/jacksontj/dataman/src/storage_node/metadata"
)

func NewStaticMetadataStore(meta *metadata.Meta) *StaticMetadataStore {
	return &StaticMetadataStore{
		m: meta,
	}
}

type StaticMetadataStore struct {
	m *metadata.Meta
	l sync.RWMutex
}

// Our methods
func (s *StaticMetadataStore) SetMeta(m *metadata.Meta) {
	s.l.Lock()
	defer s.l.Unlock()
	s.m = m
}

// Interface methods

func (s *StaticMetadataStore) GetMeta(ctx context.Context) (*metadata.Meta, error) {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.m, nil
}
