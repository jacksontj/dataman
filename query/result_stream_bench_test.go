package query

import (
	"context"
	"testing"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
	"github.com/jacksontj/dataman/stream/local"
)

func BenchmarkMergeResultStreams(b *testing.B) {
	ctx := context.Background()
	streams := 3

	vals := make([]record.Record, streams)
	resultStreams := make([]*ResultStream, streams)

	for i := 0; i < streams; i++ {
		vals[i] = record.Record{"a": i}
	}

	b.ResetTimer()

	for x:=0; x<b.N; x++ {
		for i := 0; i < streams; i++ {
			resultStreams[i] = resultStreamGenerator(vals[i], 10)
		}

		resultsChan := make(chan stream.Result, 1)
		errorChan := make(chan error, 1)

		serverStream := local.NewServerStream(resultsChan, errorChan)
		clientStream := local.NewClientStream(resultsChan, errorChan)

		go MergeResultStreams(ctx, QueryArgs{Sort: []string{"a"}}, []string{"a"}, resultStreams, serverStream)

		for {
			if _, err := clientStream.Recv(); err != nil {
				break
			}
		}
	}
}
