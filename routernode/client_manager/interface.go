package clientmanager

import "github.com/jacksontj/dataman/routernode/metadata"
import "github.com/jacksontj/dataman/client"

type ClientManager interface {
	GetClient(*metadata.DatasourceInstance) (*datamanclient.Client, error)
}
