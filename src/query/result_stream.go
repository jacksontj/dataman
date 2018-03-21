package query

import (
	"fmt"
	"io"
	"strings"

	"github.com/jacksontj/dataman/src/router_node/sharding"
	"github.com/jacksontj/dataman/src/stream"
)

// Function to transform a given record
type ResultStreamItemTransformation func(map[string]interface{}) error

// Encapsulate a streaming result from the datastore
type ResultStream struct {
	// ClientStream -- this is the actual data that is coming back from the DB
	Stream stream.ClientStream `json:"-"`
	// TODO: do we need?
	Errors []string `json:"errors,omitempty"`
	// TODO: does this make sens in the result itself?
	Meta map[string]interface{} `json:"meta,omitempty"`

	started bool

	// TODO: add lock? and/or disallow changes after first item (probably better)
	// TODO: move into stream library? This is probably generally useful
	transformations []ResultStreamItemTransformation
}

func (r *ResultStream) AddTransformation(t ResultStreamItemTransformation) error {
	if r.started {
		return fmt.Errorf("cannot add transformation after stream has started consuming")
	}
	if r.transformations == nil {
		r.transformations = []ResultStreamItemTransformation{t}
	} else {
		r.transformations = append(r.transformations, t)
	}
	return nil
}

func (r *ResultStream) Recv() (map[string]interface{}, error) {
	// TODO: check r.Errors?
	result, err := r.Stream.Recv()
	if err != nil {
		return nil, err
	}

	if record, ok := result.(map[string]interface{}); ok {
		r.started = true
		// Apply Transformations
		for _, t := range r.transformations {
			if err := t(record); err != nil {
				return record, err
			}
		}
		return record, nil
	} else {
		return nil, fmt.Errorf("Invalid type on resultStream")
	}
}

func (r *ResultStream) Close() error {
	return r.Stream.Close()
}

// TODO: take context
func MergeResultStreams(pkeyFields []string, vshardResults chan *ResultStream, resultStream stream.ServerStream) {
	defer resultStream.Close()

	pkeyFieldParts := make([][]string, len(pkeyFields))
	for i, pkeyField := range pkeyFields {
		pkeyFieldParts[i] = strings.Split(pkeyField, ".")
	}

	// We want to make sure we don't duplicate return entries
	ids := make(map[uint64]struct{})

	// TODO: select, so we can have a context with a timeout and cancel
	for vshardResultStream := range vshardResults {
		// TODO: better error checking
		// TODO: pass the errors up or redisbatch?
		if len(vshardResultStream.Errors) > 0 {
			continue
		}
		for {
			record, err := vshardResultStream.Recv()
			if err != nil {
				if err != io.EOF {
					resultStream.SendError(err)
				}
				// TODO: return and error all the things!
				break
			}
			pkeyFields := make([]interface{}, len(pkeyFieldParts))
			var ok bool
			for i, pkeyField := range pkeyFieldParts {
				pkeyFields[i], ok = GetValue(record, pkeyField)
				if !ok {
					// TODO: something else?
					panic("Missing pkey in response!!!")
				}
			}
			pkey, err := (sharding.HashMethod(sharding.SHA256).Get())(sharding.CombineKeys(pkeyFields))
			if err != nil {
				panic(fmt.Sprintf("MergeResult doesn't know how to hash pkey: %v", err))
			}
			if _, ok := ids[pkey]; !ok {
				resultStream.SendResult(record)
			}
		}
	}
}
