package functiondefault

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jacksontj/dataman/src/datamantype"
)

/*
Example field using this sequence function
	"sequence": {
		"name": "sequence",
		"field_type": "_int",
		"function_default": "sequence",
		"function_default_args": {
			"name": "unique_sequencename",
			"url": "http://127.0.0.1:8079/v1/sequence/"
		}
	},

*/

// TODO: add config for timeouts, batching, etc.
type sequenceConfig struct {
	// TODO: we probably want some mechanism to make sure people aren't colliding on sequences (at least at some level -- db/perms/etc.)
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Implementations
type Sequence struct {
	cfg *sequenceConfig
}

func (s *Sequence) Init(kwargs map[string]interface{}) error {
	if kwargs == nil {
		return fmt.Errorf("No args?")
	}

	buf, err := json.Marshal(kwargs)
	if err != nil {
		return err
	}
	s.cfg = &sequenceConfig{}
	if err := json.Unmarshal(buf, s.cfg); err != nil {
		return err
	}

	// TODO: check that we got everything we needed

	return nil
}

func (s *Sequence) SupportedTypes() []datamantype.DatamanType {
	return []datamantype.DatamanType{datamantype.Int}
}

func (s *Sequence) GetDefault(ctx context.Context, defaultType datamantype.DatamanType) (interface{}, error) {

	switch defaultType {
	case datamantype.Int:
		resp, err := http.Get(s.cfg.URL + "/" + s.cfg.Name)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		val, err := datamantype.DatamanType(datamantype.Int).Normalize(string(body))
		return val, err
	default:
		return nil, fmt.Errorf("Unsupported datamanType %s", defaultType)
	}
}
