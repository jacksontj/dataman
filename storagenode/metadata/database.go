package metadata

func NewDatabase(name string) *Database {
	return &Database{
		Name:           name,
		ShardInstances: make(map[string]*ShardInstance),
	}
}

type Database struct {
	ID   int64  `json:"_id,omitempty"`
	Name string `json:"name"`

	// TODO: switch from string, if anything we should use the "_id"
	ShardInstances map[string]*ShardInstance `json:"shard_instances"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (d *Database) Equal(o *Database) bool {
	return d.Name == o.Name
}
