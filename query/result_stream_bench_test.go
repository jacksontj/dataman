package query

import (
	"context"
	"io"
	"testing"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
	"github.com/jacksontj/dataman/stream/local"
)

func BenchmarkResultStream(b *testing.B) {
	val := record.Record{"a": 1}

	var err error
	for x := 0; x < b.N; x++ {
		resultStream := resultStreamGenerator(val, 1000)
		for {
			_, err = resultStream.Recv()
			if err == io.EOF {
				break
			}
		}
	}
}

func BenchmarkResultStreamTransformation(b *testing.B) {
	val := record.Record{"a": 1}

	tF := func(r record.Record) (record.Record, error) {
		r["t"] = "transformed"
		return r, nil
	}

	var err error
	for x := 0; x < b.N; x++ {
		resultStream := resultStreamGenerator(val, 1000)
		resultStream.AddTransformation(tF)
		for {
			_, err = resultStream.Recv()
			if err == io.EOF {
				break
			}
		}
	}
}

func BenchmarkMergeResultStreamsUnique(b *testing.B) {
	ctx := context.Background()
	streams := 100

	vals := make([]record.Record, streams)

	for i := 0; i < streams; i++ {
		vals[i] = record.Record{"a": i}
	}

	clientStreams := make([]stream.ClientStream, b.N)
	serverStreams := make([]stream.ServerStream, b.N)
	resultStreams := make([][]*ResultStream, b.N)

	for x := 0; x < b.N; x++ {
		resultsChan := make(chan record.Record, 1)
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
	streams := 100

	vals := make([]record.Record, streams)

	for i := 0; i < streams; i++ {
		vals[i] = record.Record{"a": i}
	}

	clientStreams := make([]stream.ClientStream, b.N)
	serverStreams := make([]stream.ServerStream, b.N)
	resultStreams := make([][]*ResultStream, b.N)

	for x := 0; x < b.N; x++ {
		resultsChan := make(chan record.Record, 1)
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
