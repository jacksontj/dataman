package clientmanager

import "github.com/jacksontj/dataman/routernode/metadata"
import "github.com/jacksontj/dataman/client"
import "github.com/jacksontj/dataman/client/http"

// TODO: connection limits
// TODO: KA
// TODO: MFU pooling
type HTTPClientManager struct {
}

func (h *HTTPClientManager) GetClient(datasourceInstance *metadata.DatasourceInstance) (*datamanclient.Client, error) {
	transport, err := datamanhttp.NewHTTPTransport(datasourceInstance.GetURL())
	if err != nil {
		return nil, err
	}

	return &datamanclient.Client{Transport: transport}, nil
}
