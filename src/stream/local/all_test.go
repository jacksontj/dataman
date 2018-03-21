package local

import (
	"log"
	"net/http"
	"testing"

	"github.com/jacksontj/dataman/src/stream"
)

// TODO: test client cancellation
func TestLocalStreams(t *testing.T) {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	f := func() (stream.ServerStream, stream.ClientStream) {

		resultsChan := make(chan stream.Result, 1)
		errorChan := make(chan error, 1)

		server := NewServerStream(resultsChan, errorChan)
		client := NewClientStream(resultsChan, errorChan)

		return server, client
	}

	stream.StreamTest(t, f)
}
