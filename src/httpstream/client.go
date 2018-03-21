package httpstream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

func NewJSONClientStream(r io.ReadCloser) ClientStream {
	stream := &JSONClientStream{
		results:   make(chan *ResultChunk),
		errorChan: make(chan error),
		r:         r,
	}

	go stream.handleStream()
	return stream
}

type JSONClientStream struct {
	results      chan *ResultChunk
	errorChan    chan error
	currentChunk *ResultChunk
	offset       int

	r io.ReadCloser
}

func (s *JSONClientStream) Close() error {
	if s.r != nil {
		return s.r.Close()
	}
	return nil
}

// Handle reading the io.Reader in chunks
func (s *JSONClientStream) handleStream() {
	defer func() {
		s.Close()
		close(s.results)
		close(s.errorChan)
	}()

	reader := bufio.NewReader(s.r)

	for {
		buf, err := reader.ReadBytes('\n')
		if err != nil {
			s.errorChan <- err
			return
		}

		chunk := &ResultChunk{}
		if e := json.Unmarshal(buf, chunk); e != nil {
			// TODO: set the other error?
			s.results <- &ResultChunk{Error: e.Error()}
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

func (s *JSONClientStream) Recv() (Result, error) {
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
						return nil, BrokenStream{}
					} else {
						return nil, err
					}
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
