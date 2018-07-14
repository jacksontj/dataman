package httpjson

import (
	"context"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/jacksontj/dataman/stream"
)

// TODO: test client cancellation
func TestJsonStreams(t *testing.T) {
	tests := []struct {
		chunkSize     int
		flushInterval time.Duration
	}{
		{1, 0},
		{2, 0},
		{10, 0},
		{100, 0},
		{1, time.Millisecond * 10},
		{2, time.Millisecond * 10},
		{10, time.Millisecond * 10},
		{100, time.Millisecond * 10},
	}

	for i, test := range tests {
		f := func(ctx context.Context) (stream.ServerStream, stream.ClientStream) {
			// make something client + server -- we just need something that acts like a socket
			// meaning that we can read and block waiting until io.EOF (instead of reading nothing
			// and immediately exiting like a bytes.Buffer)
			reader, writer := io.Pipe()

			server := NewServerStream(ctx, test.chunkSize, test.flushInterval, writer)
			client := NewClientStream(reader)

			return server, client
		}
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stream.StreamTest(t, f)
		})
	}
}
