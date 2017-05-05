package metadata

import (
	"encoding/json"
)

type MetaFunc func() *Meta

func StaticMetaFunc(jsonString string) (MetaFunc, error) {
	var meta Meta
	err := json.Unmarshal([]byte(jsonString), &meta)
	if err != nil {
		return nil, err
	}

	return func() *Meta { return &meta }, nil
}
