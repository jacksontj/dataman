package httpjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/stream"
)

func NewClientStream(r io.ReadCloser) stream.ClientStream {
	stream := &ClientStream{
		results:   make(chan *stream.ResultChunk),
		errorChan: make(chan error),
		r:         r,
	}

	go stream.handleStream()
	return stream
}

type ClientStream struct {
	results      chan *stream.ResultChunk
	errorChan    chan error
	currentChunk *stream.ResultChunk
	offset       int

	r io.ReadCloser
}

func (s *ClientStream) Close() error {
	if s.r != nil {
		return s.r.Close()
	}
	return nil
}

// Handle reading the io.Reader in chunks
func (s *ClientStream) handleStream() {
	defer func() {
		close(s.results)
		close(s.errorChan)
		s.Close()
	}()

	reader := bufio.NewReader(s.r)

	for {
		buf, err := reader.ReadBytes('\n')
		if err != nil {
			s.errorChan <- err
			return
		}

		chunk := &stream.ResultChunk{}
		if e := json.Unmarshal(buf, chunk); e != nil {
			// TODO: set the other error?
			s.results <- &stream.ResultChunk{Error: e.Error()}
			return
		} else {
			// If we got the trailer, we are done!
			if chunk.Error == "" && (chunk.Results == nil || len(chunk.Results) == 0) {
				return
			}
			s.results <- chunk
		}
	}

}

func (s *ClientStream) Recv() (record.Record, error) {
	for {
		// If we need a new chunk, get it
		if s.currentChunk == nil || (len(s.currentChunk.Results) <= s.offset) {
			// Check for an error on the chunk we just finished processing
			if s.currentChunk != nil && s.currentChunk.Error != "" {
				return nil, fmt.Errorf(s.currentChunk.Error)
			}

			var ok bool
			select {
			case err, ok := <-s.errorChan:
				if ok {
					if err == io.EOF {
						return nil, stream.BrokenStream{}
					} else {
						return nil, err
					}
				} else {
					// If the error channel closed, then we just need to continue on
					continue
				}
			case s.currentChunk, ok = <-s.results:
				if !ok {
					return nil, io.EOF
				}
				s.offset = 0
			}
		}

		// If there is something here for us to return, lets do it
		if s.offset < len(s.currentChunk.Results) {
			r := s.currentChunk.Results[s.offset]
			s.offset++
			return r, nil
		}
	}
}
