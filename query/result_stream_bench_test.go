package query

import (
	"context"
	"testing"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
	"github.com/jacksontj/dataman/stream/local"
)

func BenchmarkMergeResultStreamsUnique(b *testing.B) {
	ctx := context.Background()
	streams := 3

	vals := make([]record.Record, streams)

	for i := 0; i < streams; i++ {
		vals[i] = record.Record{"a": i}
	}

	clientStreams := make([]stream.ClientStream, b.N)
	serverStreams := make([]stream.ServerStream, b.N)
	resultStreams := make([][]*ResultStream, b.N)

	for x := 0; x < b.N; x++ {
		resultsChan := make(chan stream.Result, 1)
		errorChan := make(chan error, 1)
		serverStreams[x] = local.NewServerStream(ctx, resultsChan, errorChan)
		clientStreams[x] = local.NewClientStream(ctx, resultsChan, errorChan)

		rstreams := make([]*ResultStream, streams)
		for i := 0; i < streams; i++ {
			rstreams[i] = resultStreamGenerator(vals[i], 10)
		}
		resultStreams[x] = rstreams
	}

	b.ResetTimer()

	for x := 0; x < b.N; x++ {

		go MergeResultStreamsUnique(ctx, QueryArgs{Sort: []string{"a"}}, []string{"a"}, resultStreams[x], serverStreams[x])

		for {
			if _, err := clientStreams[x].Recv(); err != nil {
				break
			}
		}
	}
}

func BenchmarkMergeResultStreams(b *testing.B) {
	ctx := context.Background()
	streams := 3

	vals := make([]record.Record, streams)

	for i := 0; i < streams; i++ {
		vals[i] = record.Record{"a": i}
	}

	clientStreams := make([]stream.ClientStream, b.N)
	serverStreams := make([]stream.ServerStream, b.N)
	resultStreams := make([][]*ResultStream, b.N)

	for x := 0; x < b.N; x++ {
		resultsChan := make(chan stream.Result, 1)
		errorChan := make(chan error, 1)
		serverStreams[x] = local.NewServerStream(ctx, resultsChan, errorChan)
		clientStreams[x] = local.NewClientStream(ctx, resultsChan, errorChan)

		rstreams := make([]*ResultStream, streams)
		for i := 0; i < streams; i++ {
			rstreams[i] = resultStreamGenerator(vals[i], 10)
		}
		resultStreams[x] = rstreams
	}

	b.ResetTimer()

	for x := 0; x < b.N; x++ {

		go MergeResultStreams(ctx, QueryArgs{Sort: []string{"a"}}, resultStreams[x], serverStreams[x])

		for {
			if _, err := clientStreams[x].Recv(); err != nil {
				break
			}
		}
	}
}
