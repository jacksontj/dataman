package metadata

import "fmt"

func NewShardInstance(name string) *ShardInstance {
	return &ShardInstance{
		Name:        name,
		Collections: make(map[string]*Collection),
	}
}

type ShardInstance struct {
	ID       int64  `json:"_id,omitempty"`
	Name     string `json:"name"`
	Count    int64  `json:"count"`
	Instance int64  `json:"instance"`

	Collections map[string]*Collection `json:"collections"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (s *ShardInstance) Equal(o *ShardInstance) bool {
	if s.Name != o.Name {
		return false
	}

	// TODO: enforce after we embed and parse from name
	if s.Count != o.Count && false {
		return false
	}

	// TODO: enforce after we embed and parse from name
	if s.Instance != o.Instance && false {
		return false
	}

	return true
}

func (s *ShardInstance) GetNamespaceName() string {
	// TODO: magic so we can access our internal schema using something
	if s.ID < 0 {
		return s.Name
	} else {
		// TODO: use ID instead of name
		return fmt.Sprintf("%s_%d_%d", s.Name, s.Count, s.Instance)
	}
}
