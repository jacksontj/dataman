package stream

import (
	"fmt"
	"io"
	_ "net/http/pprof"
	"strconv"
	"testing"
	"time"
)

type TestResult struct {
	Foo string `json:"foo"`
}

// function to create stuff
func makeStuff(sw ServerStream, itemCount int, errOffset int, sleepStep int) {
	defer sw.Close()
	for i := 0; i < itemCount; i++ {
		time.Sleep(time.Millisecond * time.Duration(sleepStep*i))
		sw.SendResult(&TestResult{"a"})
		if errOffset == i {
			sw.SendError(fmt.Errorf("broken"))
			return
		}
	}
}

func streamResponses(s ClientStream) ([]Result, error) {
	results := make([]Result, 0)
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

type StreamPairCreator func() (ServerStream, ClientStream)

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
			server, client := c()

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
