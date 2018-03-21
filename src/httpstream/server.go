package httpstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Since the trailer is constant, we'll calculate it once for the package and re-use it
var trailer []byte

func init() {
	trailer, _ = json.Marshal(ResultChunk{Results: []Result{}})
}

/*

Flusing:
    - on count or on time

*/

func NewJSONServerStream(ctx context.Context, chunkSize int, flushInterval time.Duration, w io.Writer) ServerStream {
	sw := &JSONServerStream{
		resultsChan:   make(chan Result),
		errorChan:     make(chan error),
		chunkSize:     chunkSize,
		flushInterval: flushInterval,

		doneChan: make(chan struct{}),
	}

	// stream the responses
	go sw.doChunking(ctx, w)
	return sw
}

// The guy that actually writes things out
type JSONServerStream struct {
	resultsChan   chan Result
	errorChan     chan error
	chunkSize     int
	flushInterval time.Duration

	closed    bool
	streamErr error
	doneChan  chan struct{}

	// Lock for Interface facing activity
	serverLock sync.Mutex
}

// SendResult will send the result r or return an error.
func (s *JSONServerStream) SendResult(r Result) error {
	s.serverLock.Lock()
	defer s.serverLock.Unlock()
	if s.streamErr != nil {
		return s.streamErr
	} else {
		s.resultsChan <- r
		return nil
	}
}

// SendError will send the error err down the stream or return an error on its own
func (s *JSONServerStream) SendError(err error) error {
	s.serverLock.Lock()
	defer s.serverLock.Unlock()
	if s.streamErr != nil {
		return s.streamErr
	} else {
		s.errorChan <- err
		s.close(err) // TODO: another error? or something to wrap it?
		return nil
	}
}

// Close will close the server stream, disallowing all future sends
func (s *JSONServerStream) Close() error {
	s.serverLock.Lock()
	s.close(fmt.Errorf("Stream Closed"))
	// Unlock before waiting for background task to complete
	s.serverLock.Unlock()
	<-s.doneChan
	return nil
}

// close is an internal close method which assumes that the serverLock is held
func (s *JSONServerStream) close(err error) {
	if !s.closed {
		close(s.resultsChan)
		close(s.errorChan)
		s.closed = true
		s.streamErr = err
	}
}

// doChunking is a background goroutine function which does the actual chunking
// of the channels to the wire
func (s *JSONServerStream) doChunking(ctx context.Context, w io.Writer) {
	defer close(s.doneChan)
	// Support iowriters that are also flushers
	flusher, _ := w.(http.Flusher)

	timer := time.NewTimer(s.flushInterval)
	buf := make([]Result, s.chunkSize)
	i := 0
	flushNow := false

	flush := func() {
		b, _ := json.Marshal(ResultChunk{Results: buf[:i]})
		if _, err := w.Write(b); err != nil {
			s.close(err)
			return
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			s.close(err)
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
		i = 0
		flushNow = false
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(s.flushInterval)
	}

	flushTrailer := func() {
		if _, err := w.Write(trailer); err != nil {
			s.close(err)
			return
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			s.close(err)
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}

	flushError := func(err error) {
		b, _ := json.Marshal(ResultChunk{Results: buf[:i], Error: err.Error()})
		if _, err := w.Write(b); err != nil {
			s.close(err)
			return
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			s.close(err)
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}

	for {
		select {
		// consume the results until it closes, if it closes flush the buffer
		// and send the trailer (this closing first means no error)
		case result, ok := <-s.resultsChan:
			if !ok {
				if i > 0 {
					flush()
				}
				flushTrailer()
				return
			}
			buf[i] = result
			i++
			if i == s.chunkSize || flushNow {
				flush()
			}
		case <-timer.C:
			if i > 0 {
				flush()
			} else {
				flushNow = true
			}
		// If we get an error, lets flush that out and immediately stop all other activity
		case err, ok := <-s.errorChan:
			if ok {
				flushError(err)
				return
			}
		case <-ctx.Done():
			flushError(ctx.Err())
			return
		}
	}
}
