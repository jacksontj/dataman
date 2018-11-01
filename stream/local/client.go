package local

import (
	"context"
	"io"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
)

func NewClientStream(ctx context.Context, resultsChan chan record.Record, errorChan chan error) stream.ClientStream {
	stream := &ClientStream{
		ctx:         ctx,
		resultsChan: resultsChan,
		errorChan:   errorChan,
	}

	return stream
}

type ClientStream struct {
	ctx         context.Context
	resultsChan chan record.Record
	errorChan   chan error
}

func (s *ClientStream) Close() error {
	return nil
}

func (s *ClientStream) Recv() (record.Record, error) {
	// TODO: implement this cleaner, its a bit of a mess since we want specific
	// priorities on channel reading
	for {
		// we want to get results first if we have them
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case result, ok := <-s.resultsChan:
			if ok {
				return result, nil
			} else {
				// if the result chan closed, we want to drain the errors if we have any
				select {
				case err, ok := <-s.errorChan:
					if ok {
						return nil, err
					}
				default:
				}
				return nil, io.EOF
			}
		default:
		}

		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case result, ok := <-s.resultsChan:
			if ok {
				return result, nil
			} else {
				// if the result chan closed, we want to drain the errors if we have any
				select {
				case err, ok := <-s.errorChan:
					if ok {
						return nil, err
					}
				default:
				}
				return nil, io.EOF
			}
		case err, ok := <-s.errorChan:
			if ok {
				return nil, err
			}
		}
	}
}
