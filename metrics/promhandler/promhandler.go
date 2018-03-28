package promhandler

import (
	"bufio"
	"net/http"

	"github.com/jacksontj/dataman/metrics"
)

func Handler(c metrics.Collectable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ch := make(chan metrics.MetricPoint)
		go func() {
			defer close(ch)
			c.Collect(ctx, ch)
		}()

		writer := bufio.NewWriter(w)

		for metricPoint := range ch {
			if _, err := writer.WriteString(metricPoint.String()); err != nil {
				return
			}
			writer.WriteByte('\n')
		}
		writer.Flush()
	}
}
