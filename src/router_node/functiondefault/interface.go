package functiondefault

import (
	"context"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

type FunctionDefault interface {
	// Take any number of arguments from "kwargs"
	Init(globalArgs map[string]interface{}, instanceArgs map[string]interface{}) error

	// Return a list of supported types
	SupportedTypes() []storagenodemetadata.DatamanType

	// NOTE: we allow this method to return an error because some services may require it
	// even though this is the case, all efforts should be made to avoid an error in the GetDefault
	// call-- as it will impact the data-path
	// Take the db/collection/field and the current record and return the value for the field
	// context is used to pass in timeouts etc. (since some functions will call remote services etc.)
	GetDefault(
		ctx context.Context,
		defaultType storagenodemetadata.DatamanType,
		db *metadata.Database,
		collection *metadata.Collection,
		field *storagenodemetadata.CollectionField,
		record map[string]interface{},
	) (interface{}, error)
}
