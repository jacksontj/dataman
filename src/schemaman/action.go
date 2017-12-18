package main

import (
	"fmt"

	"github.com/jacksontj/dataman/src/storage_node"
	"github.com/sirupsen/logrus"
)

type SchemamanAction string

const (
	Copy SchemamanAction = "copy"
)

type Action struct {
	Src        string                 `yaml:"src"`
	Dst        string                 `yaml:"dst"`
	Action     SchemamanAction        `yaml:"action"`
	ActionArgs map[string]interface{} `yaml:"args"`
}

func (a *Action) Execute(datasourceInstances map[string]*storagenode.DatasourceInstance) error {
	switch a.Action {
	case Copy:
		return a.executeCopy(datasourceInstances)
	default:
		return fmt.Errorf("Unknown action %s", a.Action)
	}
}

func (a *Action) executeCopy(datasourceInstances map[string]*storagenode.DatasourceInstance) error {

	srcDatasource, ok := datasourceInstances[a.Src]
	if !ok {
		return fmt.Errorf("bad src: %s", a.Src)
	}

	dstDatasource, ok := datasourceInstances[a.Dst]
	if !ok {
		return fmt.Errorf("bad dst: %s", a.Dst)
	}

	// Get src schema
	// TODO: this should pull from the metadata -- if we want to go from the source we need to import
	// it into the metadata to do the copy
	srcSchema := srcDatasource.StoreSchema.GetDatabase(a.ActionArgs["database"].(string))

	if srcSchema == nil {
		return fmt.Errorf("Unable to find database %s in src %s", a.ActionArgs["database"], a.Src)
	}

	if err := dstDatasource.EnsureDatabase(srcSchema); err != nil {
		return err
	}

	// TODO: for this to work we need to either (1) have the database be in the metadata
	// store of the src, or we'll need to have a mechanism to override it for the
	// run that we are doing
	// If we are supposed to copy data, lets do it
	if val, ok := a.ActionArgs["copy_data"]; ok && val != nil && val.(bool) {
		for _, shard := range srcSchema.ShardInstances {
			for _, collection := range shard.Collections {
				// TODO: better, we need to do it a few at a time, not like this
				srcResult := srcDatasource.Store.Filter(map[string]interface{}{
					"db":             srcSchema.Name,
					"shard_instance": shard.Name,
					"collection":     collection.Name,
				})
				// Get here only works if the metadata exists for the DB we are trying to copy
				// which isn't ideal-- we'd like to be able to override this since we are
				// doing the copy outside
				if srcResult.Error != "" {
					logrus.Fatalf("Unable to get data from source: %s", srcResult.Error)
				}

				for _, record := range srcResult.Return {
					dstResult := dstDatasource.Store.Insert(map[string]interface{}{
						"db":             srcSchema.Name,
						"shard_instance": shard.Name,
						"collection":     collection.Name,
						"record":         record,
					})
					if dstResult.Error != "" {
						logrus.Fatalf("Unable to set data to dst: %s", dstResult.Error)
					}
				}
			}
		}
	}

	return nil
}
