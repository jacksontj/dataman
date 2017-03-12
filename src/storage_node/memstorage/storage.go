package memstorage

import (
	"fmt"
	"sync"

	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
)

type Storage struct {
	// map of database -> table -> entries
	store map[string]map[string]map[string]map[string]interface{}

	l sync.RWMutex
}

func (s *Storage) Init(c map[string]interface{}) error {
	s.store = make(map[string]map[string]map[string]map[string]interface{})
	return nil
}

func (s *Storage) UpdateMeta(m *metadata.Meta) error {

	for dbName, db := range m.Databases {
		s.store[dbName] = make(map[string]map[string]map[string]interface{})
		for tableName, _ := range db.Tables {
			s.store[dbName][tableName] = make(map[string]map[string]interface{})
		}
	}
	return nil
}

// Do a single item get
func (s *Storage) Get(args query.QueryArgs) *query.Result {
	s.l.RLock()
	defer s.l.RUnlock()

	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "memstorage",
		},
	}

	db, ok := s.store[args["db"].(string)]
	if !ok {
		result.Error = "Unknown db " + args["db"].(string)
		return result
	}
	table, ok := db[args["table"].(string)]
	if !ok {
		result.Error = "Unknown table " + args["table"].(string)
		return result
	}

	if entry, ok := table[fmt.Sprintf("%v", args["id"])]; ok {
		result.Return = []map[string]interface{}{entry}
	} else {
		result.Error = "Key not found"
	}

	return result
}

func (s *Storage) Set(args query.QueryArgs) *query.Result {
	s.l.Lock()
	defer s.l.Unlock()

	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "memstorage",
		},
	}

	db, ok := s.store[args["db"].(string)]
	if !ok {
		result.Error = "Unknown db " + args["db"].(string)
		return result
	}
	table, ok := db[args["table"].(string)]
	if !ok {
		result.Error = "Unknown table " + args["table"].(string)
		return result
	}

	table[fmt.Sprintf("%v", args["id"])] = args["data"].(map[string]interface{})

	return result
}

func (s *Storage) Delete(args query.QueryArgs) *query.Result {
	s.l.Lock()
	defer s.l.Unlock()

	return nil
}

func (s *Storage) Filter(args query.QueryArgs) *query.Result {
	s.l.RLock()
	defer s.l.RUnlock()

	result := &query.Result{
		// TODO: more metadata, timings, etc. -- probably want config to determine
		// what all we put in there
		Meta: map[string]interface{}{
			"datasource": "memstorage",
		},
	}
	return result
}
