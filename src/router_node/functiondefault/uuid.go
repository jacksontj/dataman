package functiondefault

import (
	"context"
	"fmt"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"

	uuid "github.com/nu7hatch/gouuid"
)

// Implementations
type UUID4 struct{}

func (u *UUID4) Init(kwargs map[string]interface{}, instanceArgs map[string]interface{}) error {
	return nil
}

func (u *UUID4) SupportedTypes() []storagenodemetadata.DatamanType {
	return []storagenodemetadata.DatamanType{storagenodemetadata.String, storagenodemetadata.Text}
}

func (u *UUID4) GetDefault(ctx context.Context,
	defaultType storagenodemetadata.DatamanType,
	db *metadata.Database,
	collection *metadata.Collection,
	field *storagenodemetadata.CollectionField,
	record map[string]interface{}) (interface{}, error) {
	val, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	switch defaultType {
	case storagenodemetadata.String, storagenodemetadata.Text:
		return val.String(), nil
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
