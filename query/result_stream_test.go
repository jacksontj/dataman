package query

import (
	"reflect"
	"testing"

	"github.com/jacksontj/dataman/stream"
	"github.com/jacksontj/dataman/stream/local"
)

func resultStreamGenerator(val interface{}, count int) *ResultStream {
	resultsChan := make(chan stream.Result, 1)
	errorChan := make(chan error, 1)

	serverStream := local.NewServerStream(resultsChan, errorChan)
	clientStream := local.NewClientStream(resultsChan, errorChan)

	go func() {
		defer serverStream.Close()
		for i := 0; i < count; i++ {
			serverStream.SendResult(val)
		}
	}()

	return &ResultStream{
		Stream: clientStream,
	}
}

func TestBasic(t *testing.T) {
	val := map[string]interface{}{"a": 1}
	resultStream := resultStreamGenerator(val, 1)
	for {
		if record, err := resultStream.Recv(); err != nil {
			t.Fatalf("error: %v", err)
		} else {
			if !reflect.DeepEqual(record, val) {
				t.Fatalf("mismatch in value")
			} else {
				break
			}
		}
	}
}

func TestBasicTransformation(t *testing.T) {
	transformed := false
	tF := func(r *map[string]interface{}) error {
		(*r)["t"] = "transformed"
		transformed = true
		return nil
	}

	val := map[string]interface{}{"a": 1}
	resultStream := resultStreamGenerator(val, 1)

	resultStream.AddTransformation(tF)

	for {
		if record, err := resultStream.Recv(); err != nil {
			t.Fatalf("error: %v", err)
		} else {
			if !reflect.DeepEqual(record, val) {
				t.Fatalf("mismatch in value")
			} else {
				break
			}
		}
	}

	if !transformed {
		t.Fatalf("transform not called")
	}
}
