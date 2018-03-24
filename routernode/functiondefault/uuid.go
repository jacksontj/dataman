package functiondefault

import (
	"context"
	"fmt"

	"github.com/jacksontj/dataman/datamantype"

	uuid "github.com/nu7hatch/gouuid"
)

// Implementations
type UUID4 struct{}

func (u *UUID4) Init(kwargs map[string]interface{}) error {
	return nil
}

func (u *UUID4) SupportedTypes() []datamantype.DatamanType {
	return []datamantype.DatamanType{datamantype.String, datamantype.Text}
}

func (u *UUID4) GetDefault(ctx context.Context,
	defaultType datamantype.DatamanType,
) (interface{}, error) {
	val, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	switch defaultType {
	case datamantype.String, datamantype.Text:
		return val.String(), nil
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
