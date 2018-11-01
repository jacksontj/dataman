package stream

import "github.com/jacksontj/dataman/record"

type ClientStream interface {
	Recv() (record.Record, error)
	// Close completes receiving of items
	Close() error
}

type ServerStream interface {
	// Methods to send items
	SendResult(record.Record) error
	SendError(error) error
	// Close completes sending of items-- no more items can be sent after Close() is called
	Close() error
}
