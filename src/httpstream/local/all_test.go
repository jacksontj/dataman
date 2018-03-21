package local

import (
	"log"
	"net/http"
	"testing"

	"github.com/jacksontj/dataman/src/httpstream"
)

// TODO: test client cancellation
func TestLocalStreams(t *testing.T) {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	f := func() (httpstream.ServerStream, httpstream.ClientStream) {

		resultsChan := make(chan httpstream.Result, 1)
		errorChan := make(chan error, 1)

		server := NewServerStream(resultsChan, errorChan)
		client := NewClientStream(resultsChan, errorChan)

		return server, client
	}

	httpstream.StreamTest(t, f)
}
