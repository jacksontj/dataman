package local

import (
	"context"
	"testing"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
)

// TODO: test client cancellation
func TestLocalStreams(t *testing.T) {
	f := func(ctx context.Context) (stream.ServerStream, stream.ClientStream) {

		resultsChan := make(chan record.Record, 1)
		errorChan := make(chan error, 1)

		server := NewServerStream(ctx, resultsChan, errorChan)
		client := NewClientStream(ctx, resultsChan, errorChan)

		return server, client
	}

	stream.StreamTest(t, f)
}
