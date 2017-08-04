package clientmanager

import "github.com/jacksontj/dataman/src/router_node/metadata"
import "github.com/jacksontj/dataman/src/client"

type ClientManager interface {
	GetClient(*metadata.DatasourceInstance) (*datamanclient.Client, error)
}
