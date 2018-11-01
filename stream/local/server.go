package local

import (
	"context"
	"fmt"
	"sync"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
)

// TODO: context
func NewServerStream(ctx context.Context, resultsChan chan record.Record, errorChan chan error) stream.ServerStream {
	sw := &ServerStream{
		ctx:         ctx,
		resultsChan: resultsChan,
		errorChan:   errorChan,
		doneChan:    make(chan struct{}),
	}

	return sw
}

// The guy that actually writes things out
type ServerStream struct {
	ctx context.Context

	resultsChan chan record.Record
	errorChan   chan error

	closed    bool
	streamErr error
	doneChan  chan struct{}

	// Lock for Interface facing activity
	serverLock sync.Mutex
}

// SendResult will send the result r or return an error.
func (s *ServerStream) SendResult(r record.Record) error {
	s.serverLock.Lock()
	defer s.serverLock.Unlock()
	if s.streamErr != nil {
		return s.streamErr
	} else {
		select {
		case s.resultsChan <- r:
			return nil
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}
}

// SendError will send the error err down the stream or return an error on its own
func (s *ServerStream) SendError(err error) error {
	s.serverLock.Lock()
	defer s.serverLock.Unlock()
	if s.streamErr != nil {
		return s.streamErr
	} else {
		select {
		case s.errorChan <- err:
			s.close(err) // TODO: another error? or something to wrap it?
			return nil
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}
}

// Close will close the server stream, disallowing all future sends
func (s *ServerStream) Close() error {
	s.serverLock.Lock()
	s.close(fmt.Errorf("Stream Closed"))
	// Unlock before waiting for background task to complete
	s.serverLock.Unlock()
	select {
	case <-s.doneChan:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// close is an internal close method which assumes that the serverLock is held
func (s *ServerStream) close(err error) {
	if !s.closed {
		close(s.resultsChan)
		close(s.errorChan)
		close(s.doneChan)
		s.closed = true
		s.streamErr = err
	}
}
