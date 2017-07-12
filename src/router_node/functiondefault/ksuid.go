package functiondefault

import (
	"context"
	"fmt"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
	"github.com/segmentio/ksuid"
)

// Implementations
type KSUID struct{}

func (u *KSUID) Init(kwargs map[string]interface{}, instanceArgs map[string]interface{}) error {
	return nil
}

func (u *KSUID) SupportedTypes() []storagenodemetadata.DatamanType {
	return []storagenodemetadata.DatamanType{storagenodemetadata.String, storagenodemetadata.Text}
}

func (u *KSUID) GetDefault(ctx context.Context,
	defaultType storagenodemetadata.DatamanType,
	db *metadata.Database,
	collection *metadata.Collection,
	field *storagenodemetadata.CollectionField,
	record map[string]interface{}) (interface{}, error) {

	switch defaultType {
	case storagenodemetadata.String, storagenodemetadata.Text:
		val, err := ksuid.NewRandom()
		if err != nil {
			return val, err
		}
		return val.String(), nil
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
