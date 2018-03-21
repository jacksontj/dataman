package routernode

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/pprof"

	"github.com/julienschmidt/httprouter"

	"github.com/jacksontj/dataman/src/httputil"
	"github.com/jacksontj/dataman/src/query"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
)

type HTTPApi struct {
	routerNode *RouterNode
}

func NewHTTPApi(routerNode *RouterNode) *HTTPApi {
	api := &HTTPApi{
		routerNode: routerNode,
	}

	return api
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return httputil.LoggingHandler(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.ServeHTTP(w, r)
	})
}

// Register any endpoints to the router
func (h *HTTPApi) Start(router *httprouter.Router) {
	// Just dump the current meta we have
	router.GET("/v1/metadata", httputil.LoggingHandler(h.showMetadata))

	// Data access APIs
	router.POST("/v1/data/raw/:qtype", httputil.LoggingHandler(h.rawQueryHandler))

	// TODO: options to enable/disable (or scope to just localhost)
	router.GET("/v1/debug/pprof/", wrapHandler(http.HandlerFunc(pprof.Index)))
	router.GET("/v1/debug/pprof/cmdline", wrapHandler(http.HandlerFunc(pprof.Cmdline)))
	router.GET("/v1/debug/pprof/profile", wrapHandler(http.HandlerFunc(pprof.Profile)))
	router.GET("/v1/debug/pprof/symbol", wrapHandler(http.HandlerFunc(pprof.Symbol)))
	router.GET("/v1/debug/pprof/trace", wrapHandler(http.HandlerFunc(pprof.Trace)))

	// TODO: wrap a different registry (if we ever want more than one per process)
	router.GET("/v1/debug/metrics", wrapHandler(exp.ExpHandler(metrics.DefaultRegistry)))
}

// List all databases that we have in the metadata store
func (h *HTTPApi) showMetadata(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	meta := h.routerNode.GetMeta()

	// Now we need to return the results
	if bytes, err := json.Marshal(meta); err != nil {
		// TODO: log this better?
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

// TODO: streaming parser
/*

	API requests on the router do the following
		- database
		- datasource
		- shard
		- shard item (pick the replica)
		- forward to storage_node

*/
// TODO: implement
func (h *HTTPApi) rawQueryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	defer r.Body.Close()
	bytes, _ := ioutil.ReadAll(r.Body)

	// TODO: validate that this is correct, error if its not a valid name
	qType := query.QueryType(ps.ByName("qtype"))

	var qArgs query.QueryArgs

	if err := json.Unmarshal(bytes, &qArgs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else {
		// Otherwise, lets create the query struct to pass down
		q := query.Query{
			Type: qType,
			Args: qArgs,
		}

		results := h.routerNode.HandleQuery(ctx, &q)
		// Now we need to return the results
		if bytes, err := json.Marshal(results); err != nil {
			// TODO: log this better?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	}
}
