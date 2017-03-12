package storagenode

import (
	"github.com/jacksontj/dataman/src/metadata"
	"github.com/jacksontj/dataman/src/query"
)

// Interface that a storage node must implement
type StorageInterface interface {
	// Initialization, this is the "config_json" for the `storage_node`
	Init(map[string]interface{}) error

	// Schema-Functions
	// AddDatabase
	// ListDatabase
	// RemoveDatabase
	// AddTable
	// ListTable
	// RemoveTable
	// AddIndex
	// ListIndex
	// RemoveIndex
	// TODO: replace this with simpler methods, this update mechanism is probably generic
	// and we'll just need add/remove/list/update methods to accomplish the task
	UpdateMeta(*metadata.Meta) error

	// Data-Functions
	// TODO: split out the various functions into grouped interfaces that make sense
	// for now we'll just have one, but eventually we could support "TransactionalStorageNode" etc.
	// TODO: more specific types for each method
	Get(query.QueryArgs) *query.Result
	Set(query.QueryArgs) *query.Result
	//Delete(query.QueryArgs) *query.Result
	Filter(query.QueryArgs) *query.Result
}
