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
