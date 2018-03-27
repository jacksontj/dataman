package promhandler

import (
	"bufio"
	"fmt"
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
			_, err := writer.WriteString(metricPoint.String())
			// TODO: do something with the error
			fmt.Println("err", err)
			writer.WriteByte('\n')
		}
		writer.Flush()
	}
}
