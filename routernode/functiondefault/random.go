package functiondefault

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jacksontj/dataman/datamantype"
)

// Implementations
type Random struct{}

func (u *Random) Init(kwargs map[string]interface{}) error {
	return nil
}

func (u *Random) SupportedTypes() []datamantype.DatamanType {
	return []datamantype.DatamanType{datamantype.Int}
}

func (u *Random) GetDefault(ctx context.Context,
	defaultType datamantype.DatamanType,
) (interface{}, error) {

	switch defaultType {
	case datamantype.Int:
		return rand.Int(), nil
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
