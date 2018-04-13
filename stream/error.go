package stream

// BrokenStream means that the stream broke before we got the trailer.
// In practice this means that there was an interruption to the stream, meaning
// the results are not complete-- but all previously received results before
// the erorr are still valid they just don't constitute the full set of results
type BrokenStream struct{}

func (b BrokenStream) Error() string { return "premature end of stream" }
