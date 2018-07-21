package query

import (
	"container/heap"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jacksontj/dataman/record"
	"github.com/jacksontj/dataman/routernode/sharding"
	"github.com/jacksontj/dataman/stream"
)

// Function to transform a given record
// argument *must* be a pointer, otherwise you are passed a copy which you cannot replace directly
type ResultStreamItemTransformation func(record.Record) (record.Record, error)

// Encapsulate a streaming result from the datastore
type ResultStream struct {
	// ClientStream -- this is the actual data that is coming back from the DB
	Stream stream.ClientStream `json:"-"`
	// TODO: do we need?
	Errors []string `json:"errors,omitempty"`
	// TODO: does this make sens in the result itself?
	Meta map[string]interface{} `json:"meta,omitempty"`

	started bool

	// TODO: add lock? and/or disallow changes after first item (probably better)
	// TODO: move into stream library? This is probably generally useful
	transformations []ResultStreamItemTransformation
}

func (r *ResultStream) Err() error {
	if r.Errors == nil {
		return nil
	} else {
		return fmt.Errorf(strings.Join(r.Errors, "\n"))
	}
}

func (r *ResultStream) AddTransformation(t ResultStreamItemTransformation) error {
	if r.started {
		return fmt.Errorf("cannot add transformation after stream has started consuming")
	}
	if r.transformations == nil {
		r.transformations = []ResultStreamItemTransformation{t}
	} else {
		r.transformations = append(r.transformations, t)
	}
	return nil
}

func (r *ResultStream) Recv() (record.Record, error) {
	// TODO: check r.Errors?
	result, err := r.Stream.Recv()
	if err != nil {
		return nil, err
	}

	var resultRecord record.Record
	// The type switch here is necessary as the stream type (as of now) is
	// an interface{} -- so things that are unmarshalled (for example from json)
	// might come in as a generic map[string]interface{}
	switch resultTyped := result.(type) {
	case map[string]interface{}:
		resultRecord = record.Record(resultTyped)
	case record.Record:
		resultRecord = resultTyped
	default:
		return nil, fmt.Errorf("Invalid type on resultStream")
	}

	r.started = true
	// Apply Transformations
	for _, t := range r.transformations {
		if transformedRecord, err := t(resultRecord); err != nil {
			return resultRecord, err
		} else if transformedRecord != nil {
			resultRecord = transformedRecord
		}
	}
	return resultRecord, nil
}

func (r *ResultStream) Close() error {
	if r.Stream != nil {
		return r.Stream.Close()
	}
	return nil
}

// TODO: move to stream?
type streamItem struct {
	item record.Record
	err  error
}

func streamResults(ctx context.Context, stream *ResultStream) chan *streamItem {
	// TODO: configurable size?
	c := make(chan *streamItem, 1000)
	go func(stream *ResultStream) {
		defer close(c)
		for {
			v, err := stream.Recv()
			if err == io.EOF {
				return
			}
			select {
			case c <- &streamItem{v, err}:
			// If the context is closed, then we need to cancel, we'll do a
			// non-blocking send of an error down the channel, and then exit
			case <-ctx.Done():
				select {
				case c <- &streamItem{err: ctx.Err()}:
				default:
				}
				return
			}
		}
	}(stream)
	return c
}

// TODO: cleaner? seems that this is faster than the reflect one
// TODO: move to stream package?
func mergeStreams(ctx context.Context, streams []*ResultStream) chan streamItem {
	// TODO: configurable size?
	c := make(chan streamItem, 1000)
	wg := &sync.WaitGroup{}
	for _, stream := range streams {
		wg.Add(1)
		go func(stream *ResultStream) {
			defer wg.Done()
			for {
				v, err := stream.Recv()
				if err == io.EOF {
					return
				}
				select {
				case c <- streamItem{v, err}:
				// If the context is closed, then we need to cancel, we'll do a
				// non-blocking send of an error down the channel, and then exit
				case <-ctx.Done():
					select {
					case c <- streamItem{err: ctx.Err()}:
					default:
					}
					return
				}
			}
		}(stream)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	return c
}

// MergeResultStreams is responsible to (1) merge result streams uniquely based on pkey and (2) maintain sort order
// (if sorted) from the streams (each stream is assumed to be in-order already)
func MergeResultStreamsUnique(ctx context.Context, args QueryArgs, pkeyFields []string, vshardResults []*ResultStream, resultStream stream.ServerStream) {
	defer resultStream.Close()

	pkeyFieldParts := make([][]string, len(pkeyFields))
	for i, pkeyField := range pkeyFields {
		pkeyFieldParts[i] = strings.Split(pkeyField, ".")
	}

	getPkeyID := func(r record.Record) uint64 {
		// now get the pkey from the item, to ensure no dupes
		pkeyFields := make([]interface{}, len(pkeyFieldParts))
		var ok bool
		for i, pkeyField := range pkeyFieldParts {
			pkeyFields[i], ok = r.Get(pkeyField)
			if !ok {
				// TODO: something else?
				panic("Missing pkey in response!!!")
			}
		}
		pkey, err := (sharding.HashMethod(sharding.SHA256).Get())(sharding.CombineKeys(pkeyFields))
		if err != nil {
			panic(fmt.Sprintf("MergeResult doesn't know how to hash pkey: %v", err))
		}
		return pkey
	}

	// We want to make sure we don't duplicate return entries
	ids := make(map[uint64]struct{})
	offset := args.Offset

	// If we need to do sorting, then we need to do a minheap thing
	if args.Sort != nil {
		// create slice of stream channels
		vshardResultChannels := make([]chan *streamItem, len(vshardResults))
		for i, vshardResult := range vshardResults {
			defer vshardResult.Close()

			if vshardResult.Errors != nil && len(vshardResult.Errors) > 0 {
				resultStream.SendError(fmt.Errorf("errors in thing: %v", vshardResult.Errors))
				return
			}
			if vshardResult.Stream == nil {
				resultStream.SendError(fmt.Errorf("no stream in resultStream"))
				return
			}
			vshardResultChannels[i] = streamResults(ctx, vshardResult)
		}

		if args.SortReverse == nil {
			sortReverseList := make([]bool, len(args.Sort))
			// TODO: better, seems heavy
			for i := range sortReverseList {
				sortReverseList[i] = false
			}
			args.SortReverse = sortReverseList
		}

		// TODO: util func
		splitSortKeys := make([][]string, len(args.Sort))
		for i, sortKey := range args.Sort {
			splitSortKeys[i] = strings.Split(sortKey, ".")
		}

		// TODO: refactor
		h := record.NewRecordHeap(splitSortKeys, args.SortReverse)
		heap.Init(h)

		// Load an item from each vshard
		for i, source := range vshardResultChannels {
			if head, ok := <-source; ok {
				if head.err != nil {
					resultStream.SendError(head.err)
					return
				}
				item := record.RecordItem{
					Record: head.item,
					Source: i,
				}
				heap.Push(h, item)
			}
		}

		// TODO: faster to just use len(ids) ?
		// Count of the number of results we've sent
		resultsSent := uint64(0)

		for h.Len() > 0 {
			item := heap.Pop(h).(record.RecordItem)

			// now get the pkey from the item, to ensure no dupes
			pkeyID := getPkeyID(item.Record)
			if _, ok := ids[pkeyID]; !ok {
				ids[pkeyID] = struct{}{}

				// If an offset was defined, do that
				if offset > 0 {
					offset--
				} else {
					if err := resultStream.SendResult(item.Record); err != nil {
						resultStream.SendError(err)
						return
					}
					resultsSent++
					// If we have a limit defined, lets enforce it
					if args.Limit > 0 && resultsSent >= args.Limit {
						return
					}
				}
			}

			// TODO: add item back
			source := vshardResultChannels[item.Source]
			if head, ok := <-source; ok {
				if head.err != nil {
					resultStream.SendError(head.err)
					return
				}
				newItem := record.RecordItem{
					Record: head.item,
					Source: item.Source,
				}
				heap.Push(h, newItem)
			}
		}
	} else {
		// TODO: faster to just use len(ids) ?
		// Count of the number of results we've sent
		resultsSent := uint64(0)

		// TODO: move into the mergeStreams method?
		// This is ugly, but we need to check that these don't have errors, and if we do we want to propagate them
		for _, vshardResult := range vshardResults {
			// TODO: benchmark 100s of defers -- if this is slow move it to a function as a single defer
			defer vshardResult.Close()

			if vshardResult.Errors != nil && len(vshardResult.Errors) > 0 {
				resultStream.SendError(fmt.Errorf("errors in thing: %v", vshardResult.Errors))
				return
			}
			if vshardResult.Stream == nil {
				resultStream.SendError(fmt.Errorf("no stream in resultStream"))
				return
			}
		}

		// Othewise we just need to select through the channels and push results as we get them
		s := mergeStreams(ctx, vshardResults)
		for item := range s {
			if item.err != nil {
				resultStream.SendError(item.err)
				return
			} else {
				// now get the pkey from the item, to ensure no dupes
				pkeyID := getPkeyID(item.item)
				if _, ok := ids[pkeyID]; !ok {
					ids[pkeyID] = struct{}{}

					// If an offset was defined, do that
					if offset > 0 {
						offset--
					} else {
						if err := resultStream.SendResult(item.item); err != nil {
							resultStream.SendError(err)
							return
						}
						resultsSent++
						// If we have a limit defined, lets enforce it
						if args.Limit > 0 && resultsSent >= args.Limit {
							return
						}
					}
				}
			}
		}
	}
}

// TODO: take context
// MergeResultStreams is responsible to (1) merge result streams and (2) maintain sort order
// (if sorted) from the streams (each stream is assumed to be in-order already)
func MergeResultStreams(ctx context.Context, args QueryArgs, vshardResults []*ResultStream, resultStream stream.ServerStream) {
	defer resultStream.Close()

	// We want to make sure we don't duplicate return entries
	offset := args.Offset

	// If we need to do sorting, then we need to do a minheap thing
	if args.Sort != nil {
		// create slice of stream channels
		vshardResultChannels := make([]chan *streamItem, len(vshardResults))
		for i, vshardResult := range vshardResults {
			defer vshardResult.Close()

			if vshardResult.Errors != nil && len(vshardResult.Errors) > 0 {
				resultStream.SendError(fmt.Errorf("errors in thing: %v", vshardResult.Errors))
				return
			}
			if vshardResult.Stream == nil {
				resultStream.SendError(fmt.Errorf("no stream in resultStream"))
				return
			}
			vshardResultChannels[i] = streamResults(ctx, vshardResult)
		}

		if args.SortReverse == nil {
			sortReverseList := make([]bool, len(args.Sort))
			// TODO: better, seems heavy
			for i := range sortReverseList {
				sortReverseList[i] = false
			}
			args.SortReverse = sortReverseList
		}

		// TODO: util func
		splitSortKeys := make([][]string, len(args.Sort))
		for i, sortKey := range args.Sort {
			splitSortKeys[i] = strings.Split(sortKey, ".")
		}

		// TODO: refactor
		h := record.NewRecordHeap(splitSortKeys, args.SortReverse)
		heap.Init(h)

		// Load an item from each vshard
		for i, source := range vshardResultChannels {
			if head, ok := <-source; ok {
				if head.err != nil {
					resultStream.SendError(head.err)
					return
				}
				item := record.RecordItem{
					Record: head.item,
					Source: i,
				}
				heap.Push(h, item)
			}
		}

		// TODO: faster to just use len(ids) ?
		// Count of the number of results we've sent
		resultsSent := uint64(0)

		for h.Len() > 0 {
			item := heap.Pop(h).(record.RecordItem)

			// now get the pkey from the item, to ensure no dupes
			// If an offset was defined, do that
			if offset > 0 {
				offset--
			} else {
				if err := resultStream.SendResult(item.Record); err != nil {
					resultStream.SendError(err)
					return
				}
				resultsSent++
				// If we have a limit defined, lets enforce it
				if args.Limit > 0 && resultsSent >= args.Limit {
					return
				}
			}

			// TODO: add item back
			source := vshardResultChannels[item.Source]
			if head, ok := <-source; ok {
				if head.err != nil {
					resultStream.SendError(head.err)
					return
				}
				newItem := record.RecordItem{
					Record: head.item,
					Source: item.Source,
				}
				heap.Push(h, newItem)
			}
		}
	} else {
		// TODO: faster to just use len(ids) ?
		// Count of the number of results we've sent
		resultsSent := uint64(0)

		// TODO: move into the mergeStreams method?
		// This is ugly, but we need to check that these don't have errors, and if we do we want to propagate them
		for _, vshardResult := range vshardResults {
			// TODO: benchmark 100s of defers -- if this is slow move it to a function as a single defer
			defer vshardResult.Close()

			if vshardResult.Errors != nil && len(vshardResult.Errors) > 0 {
				resultStream.SendError(fmt.Errorf("errors in thing: %v", vshardResult.Errors))
				return
			}
			if vshardResult.Stream == nil {
				resultStream.SendError(fmt.Errorf("no stream in resultStream"))
				return
			}
		}

		// Othewise we just need to select through the channels and push results as we get them
		s := mergeStreams(ctx, vshardResults)
		for item := range s {
			if item.err != nil {
				resultStream.SendError(item.err)
				return
			} else {
				// If an offset was defined, do that
				if offset > 0 {
					offset--
				} else {
					if err := resultStream.SendResult(item.item); err != nil {
						resultStream.SendError(err)
						return
					}
					resultsSent++
					// If we have a limit defined, lets enforce it
					if args.Limit > 0 && resultsSent >= args.Limit {
						return
					}
				}
			}
		}
	}
}
