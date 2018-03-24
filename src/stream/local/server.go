package local

import (
	"fmt"
	"sync"

	"github.com/jacksontj/dataman/src/stream"
)

// TODO: context
func NewServerStream(resultsChan chan stream.Result, errorChan chan error) stream.ServerStream {
	sw := &ServerStream{
		resultsChan: resultsChan,
		errorChan:   errorChan,
		doneChan:    make(chan struct{}),
	}

	return sw
}

// The guy that actually writes things out
type ServerStream struct {
	resultsChan chan stream.Result
	errorChan   chan error

	closed    bool
	streamErr error
	doneChan  chan struct{}

	// Lock for Interface facing activity
	serverLock sync.Mutex
}

// SendResult will send the result r or return an error.
func (s *ServerStream) SendResult(r stream.Result) error {
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
func (s *ServerStream) SendError(err error) error {
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
func (s *ServerStream) Close() error {
	s.serverLock.Lock()
	s.close(fmt.Errorf("Stream Closed"))
	// Unlock before waiting for background task to complete
	s.serverLock.Unlock()
	<-s.doneChan
	return nil
}

// close is an internal close method which assumes that the serverLock is held
func (s *ServerStream) close(err error) {
	if !s.closed {
		close(s.resultsChan)
		close(s.errorChan)
		s.closed = true
		s.streamErr = err
	}
}
