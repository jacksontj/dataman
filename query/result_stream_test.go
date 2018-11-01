package query

import (
	"context"
	"reflect"
	"testing"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream/local"
)

func resultStreamGenerator(val record.Record, count int) *ResultStream {
	ctx := context.Background()
	resultsChan := make(chan record.Record, 1)
	errorChan := make(chan error, 1)

	serverStream := local.NewServerStream(ctx, resultsChan, errorChan)
	clientStream := local.NewClientStream(ctx, resultsChan, errorChan)

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
	val := record.Record{"a": 1}
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
	tF := func(r record.Record) (record.Record, error) {
		r["t"] = "transformed"
		transformed = true
		return r, nil
	}

	val := record.Record{"a": 1}
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
