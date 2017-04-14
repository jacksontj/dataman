package metadata

import (
	"encoding/json"
	"io/ioutil"
)

type MetaFunc func() *Meta

func StaticMetaFunc(filepath string) (MetaFunc, error) {
	var meta Meta
	configBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(configBytes), &meta)
	if err != nil {
		return nil, err
	}

	// TODO: remove this, really need to do this elsewhere
	for _, database := range meta.Databases {
		for _, collection := range database.Collections {
			collection.FieldMap = make(map[string]*Field)
			for _, field := range collection.Fields {
				collection.FieldMap[field.Name] = field
			}
		}
	}

	return func() *Meta { return &meta }, nil
}
