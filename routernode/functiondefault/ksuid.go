package functiondefault

import (
	"context"
	"fmt"

	"github.com/jacksontj/dataman/datamantype"
	"github.com/segmentio/ksuid"
)

// Implementations
type KSUID struct{}

func (u *KSUID) Init(kwargs map[string]interface{}) error {
	return nil
}

func (u *KSUID) SupportedTypes() []datamantype.DatamanType {
	return []datamantype.DatamanType{datamantype.String, datamantype.Text}
}

func (u *KSUID) GetDefault(ctx context.Context,
	defaultType datamantype.DatamanType,
) (interface{}, error) {

	switch defaultType {
	case datamantype.String, datamantype.Text:
		val, err := ksuid.NewRandom()
		if err != nil {
			return val, err
		}
		return val.String(), nil
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
