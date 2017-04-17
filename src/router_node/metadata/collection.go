package metadata

func NewCollection(name string) *Collection {
	return &Collection{
		Name: name,
	}
}

type Collection struct {
	Name string

	// TODO: use, we don't need these for inital working product, but we will
	// if we plan on doing more sophisticated sharding or schema validation
	//Fields map[string]*CollectionField
	//Indexes map[string]*CollectionIndex
}

// TODO: fill out
type CollectionField struct {
	Name string
}

// TODO: fill out
type CollectionIndex struct {
	Name string
}
