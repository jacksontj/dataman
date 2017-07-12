package functiondefault

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jacksontj/dataman/src/router_node/metadata"
	storagenodemetadata "github.com/jacksontj/dataman/src/storage_node/metadata"
)

// Implementations
type Random struct{}

func (u *Random) Init(kwargs map[string]interface{}, instanceArgs map[string]interface{}) error {
	return nil
}

func (u *Random) SupportedTypes() []storagenodemetadata.DatamanType {
	return []storagenodemetadata.DatamanType{storagenodemetadata.Int}
}

func (u *Random) GetDefault(ctx context.Context,
	defaultType storagenodemetadata.DatamanType,
	db *metadata.Database,
	collection *metadata.Collection,
	field *storagenodemetadata.CollectionField,
	record map[string]interface{}) (interface{}, error) {

	switch defaultType {
	case storagenodemetadata.Int:
		return rand.Int(), nil
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
