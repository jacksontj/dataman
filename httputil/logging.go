package httputil

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// TODO: async logging to disk? (stdout is relatively slow)
func LoggingHandler(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		start := time.Now()
		h(w, r, ps)
		fmt.Println(r.Method, r.URL, time.Now().Sub(start))
	}
}
