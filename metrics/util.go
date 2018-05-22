package metrics

import "context"

// CollectOne will get a single MetricPoint from a Collectable c
func CollectOne(ctx context.Context, c Collectable) MetricPoint {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	ch := make(chan MetricPoint)

	go c.Collect(childCtx, ch)
	return <-ch
}
