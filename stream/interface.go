package stream

type ClientStream interface {
	Recv() (Result, error)
	// Close completes receiving of items
	Close() error
}

type ServerStream interface {
	// Methods to send items
	SendResult(Result) error
	SendError(error) error
	// Close completes sending of items-- no more items can be sent after Close() is called
	Close() error
}
