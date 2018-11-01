package stream

import (
	"context"
	"fmt"
	"io"
	_ "net/http/pprof"
	"strconv"
	"testing"
	"time"

	"github.com/jacksontj/dataman/record"
)

// function to create stuff
func makeStuff(sw ServerStream, itemCount int, errOffset int, sleepStep int) {
	defer sw.Close()
	for i := 0; i < itemCount; i++ {
		time.Sleep(time.Millisecond * time.Duration(sleepStep*i))
		sw.SendResult(record.Record{"foo": "a"})
		if errOffset == i {
			sw.SendError(fmt.Errorf("broken"))
			return
		}
	}
}

func streamResponses(s ClientStream) ([]record.Record, error) {
	results := make([]record.Record, 0)
	for {
		item, err := s.Recv()
		if err != nil {
			// IOF means that we are done with the stream with no error
			if err == io.EOF {
				return results, nil
			}
			return results, err
		}
		results = append(results, item)
	}
}

type StreamPairCreator func(context.Context) (ServerStream, ClientStream)

func StreamTest(t *testing.T, c StreamPairCreator) {
	tests := []struct {
		itemCount int
		errOffset int
		sleepStep int
	}{
		{2, -1, 0},
		{2, 0, 0},
		{2, 1, 0},
		{2, 2, 0},

		{2, -1, 10},
		{2, 0, 10},
		{2, 1, 10},
		{2, 2, 10},

		{10, -1, 0},
		{10, 0, 0},
		{10, 1, 0},
		{10, 10, 0},

		{10, -1, 10},
		{10, 0, 10},
		{10, 1, 10},
		{10, 10, 10},
	}

	for i, test := range tests {
		expectingError := test.errOffset >= 0 && test.errOffset < test.itemCount
		expectedResults := test.itemCount
		if test.errOffset >= 0 && test.errOffset < test.itemCount {
			expectedResults = test.errOffset + 1
		}

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			server, client := c(ctx)

			// return 10 items, no errors
			go makeStuff(server, test.itemCount, test.errOffset, test.sleepStep)

			results, err := streamResponses(client)
			if err != nil && !expectingError {
				t.Fatalf("error unexpected: %v", err)
			} else if err == nil && expectingError {
				t.Fatalf("Missing expected error")
			}
			if len(results) != expectedResults {
				t.Fatalf("incorrect number of responses expected=%d actual=%d", expectedResults, len(results))
			}
			server.Close()
		})
	}
}
