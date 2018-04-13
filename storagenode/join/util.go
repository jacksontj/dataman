package join

import (
	"fmt"
	"sort"
	"strings"

	radix "github.com/armon/go-radix"
	"github.com/jacksontj/dataman/storagenode/metadata"
)

// TODO: map[string]Filter
type JoinMap map[string]map[string]interface{}

func ParseJoinMap(joinField interface{}) (JoinMap, error) {
	joinMap := make(JoinMap)
	switch joinFieldTyped := joinField.(type) {
	case []interface{}:
		for _, item := range joinFieldTyped {
			stringKey, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("joinFieldList must be []string")
			}
			joinMap[stringKey] = make(map[string]interface{})
		}

	case []string:
		for _, item := range joinFieldTyped {
			joinMap[item] = make(map[string]interface{})
		}

	case map[string]interface{}:
		for k, v := range joinFieldTyped {
			if vTyped, ok := v.(map[string]interface{}); ok {
				joinMap[k] = vTyped
			} else {
				return nil, fmt.Errorf("values must be map[string]interface{}")
			}
		}

	default:
		// TODO: be more explicit
		return nil, fmt.Errorf("Invalid join type")
	}
	return joinMap, nil
}

func SortJoinFieldList(l []string) {
	// Sort joinFieldList shortest -> longest -- so we do the deps first
	less := func(i, j int) bool {
		return strings.Count(l[i], ".") < strings.Count(l[j], ".")
	}
	sort.Slice(l, less)
}

// Repsonsible for splitting and normalizing, not sorting
func SplitJoinFieldsString(collectionName string, joinMap JoinMap) ([]string, []string) {
	NormalizeJoinMap(collectionName, joinMap)
	forwardList := make([]string, 0)
	reverseList := make([]string, 0)
	for joinItem := range joinMap {
		if strings.HasPrefix(joinItem, ".") {
			forwardList = append(forwardList, joinItem)
		} else {
			reverseList = append(reverseList, joinItem)
		}
	}
	SortJoinFieldList(forwardList)
	SortJoinFieldList(reverseList)

	return forwardList, reverseList
}

func NormalizeJoinMap(n string, m JoinMap) {
	prefix := n + "."
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			m[strings.TrimPrefix(k, n)] = v
			delete(m, k)
		}
	}
}

// Since router and storage_node have different collections we need an interface which is common)
type MetaCollection interface {
	GetName() string
	GetFieldByName(name string) *metadata.CollectionField
	GetField(nameParts []string) *metadata.CollectionField
}

type CollectionGetter func(name string) (MetaCollection, error)

func NewCollection(m MetaCollection) *Collection {
	return &Collection{
		M:           m,
		Name:        m.GetName(),
		ForwardJoin: make([]*ForwardJoin, 0),
	}
}

// Represent a collection in the join state
type Collection struct {
	// A pointer to us-- so we can do things like find fields etc.
	M    MetaCollection `json:"-"`
	Name string         `json:"name"`

	ForwardJoin []*ForwardJoin `json:"forward_join,omitempty"`
	// There can be N reverse ones
	ReverseJoin []*ReverseJoin `json:"reverse_join,omitempty"`
}

func (c *Collection) HasJoins() bool {
	return !(len(c.ForwardJoin) == 0 && len(c.ReverseJoin) == 0)
}

type ForwardJoin struct {
	Key       string                    `json:"key"`
	C         *Collection               `json:"c"`
	JoinField *metadata.CollectionField `json:"join_field"`
	Filter    map[string]interface{}    `json:"filter"`
}

type ReverseJoin struct {
	Key       string                    `json:"key"`
	C         *Collection               `json:"c"`
	JoinField *metadata.CollectionField `json:"join_field"`
	Filter    map[string]interface{}    `json:"filter"`
}

// Return a list of join steps. These can be run in chains
func OrderJoins(collectionMetaGetter CollectionGetter, collection MetaCollection, joinMap JoinMap) (*Collection, error) {
	// Normalize all the ".<key>" to "<collection>.<key>"
	NormalizeJoinMap(collection.GetName(), joinMap)

	forwardJoinList, reverseJoinList := SplitJoinFieldsString(collection.GetName(), joinMap)

	// Map of collection -> joinCollection repr -- this is for reverse joins (basically)
	thisCollection := NewCollection(collection)
	collectionMap := map[string][]*Collection{collection.GetName(): {thisCollection}}

	forwardJoinFunc := func(thisCollection *Collection, forwardJoinList []string) error {
		// For forward joins we want to check for prefix matching and do our mapping that way
		prefixTree := radix.New()

		// Do forward joins
		for _, join := range forwardJoinList {
			joinParts := strings.Split(join, ".")
			joinParts = joinParts[1:]

			var joinField *metadata.CollectionField
			baseCollection := thisCollection
			if prefix, item, ok := prefixTree.LongestPrefix(join); ok {
				baseJoinCollection := item.(*Collection)
				baseCollection = baseJoinCollection

				suffix := strings.TrimPrefix(join, prefix+".")
				joinParts = joinParts[strings.Count(prefix, "."):]

				joinField = baseJoinCollection.M.GetFieldByName(suffix)
			} else {
				joinField = thisCollection.M.GetField(joinParts)
			}

			if joinField == nil {
				return fmt.Errorf("Invalid join key: %s", join)
			}

			// TODO: check for non-existance
			metaJoinCollection, err := collectionMetaGetter(joinField.Relation.Collection)
			if err != nil {
				return err
			}
			joinCollection := NewCollection(metaJoinCollection)

			forwardJoin := &ForwardJoin{
				Key:       strings.Join(joinParts, "."),
				C:         joinCollection,
				JoinField: joinField,
				Filter:    joinMap[join],
			}

			// TODO: we *could* support it, but it would be odd, because we
			// overwrite the joinkey later (since it has to match to join)
			// we can change this in the future to only do the lookup if the
			// key matches? (since thats presumably what people want?)
			if _, ok := forwardJoin.Filter[forwardJoin.JoinField.Relation.Field]; ok {
				return fmt.Errorf("Join filter cannot include the join field")
			}

			baseCollection.ForwardJoin = append(baseCollection.ForwardJoin, forwardJoin)

			prefixTree.Insert(join, joinCollection)
		}
		return nil
	}

	if err := forwardJoinFunc(thisCollection, forwardJoinList); err != nil {
		return nil, err
	}

	// Now do reverse ones

	// While doing reverse joins we may find that there are no more reverse to do, so we'll need to do a round
	// of forward first, then continue

	// If we find a match here then we need to add ourselves as a key to add
	// prefix -> *Collection
	joinRadixTree := radix.New()
	// map of *Collection -> joinKeys
	forwardJoinMap := make(map[*Collection][]string)

	lastLen := 0
	for {
		reverseJoinListLen := len(reverseJoinList)
		if reverseJoinListLen <= 0 {
			break
		}
		if reverseJoinListLen == lastLen {
			if len(forwardJoinMap) > 0 {
				for k, v := range forwardJoinMap {
					if err := forwardJoinFunc(k, v); err != nil {
						return nil, err
					}
				}
				forwardJoinMap = make(map[*Collection][]string)
				continue
			} else {
				return nil, fmt.Errorf("Unable to complete reverse joins")
			}
		}
		lastLen = reverseJoinListLen

		toRemove := -1
		for i, join := range reverseJoinList {

			// if it was a match, we add it to the list and continue
			if prefix, item, ok := joinRadixTree.LongestPrefix(join); ok {
				baseJoinCollection := item.(*Collection)
				joinKeys, ok := forwardJoinMap[baseJoinCollection]
				if !ok {
					joinKeys = make([]string, 0, 1)
				}
				forwardJoinMap[baseJoinCollection] = append(joinKeys, strings.TrimPrefix(join, prefix))
				// TODO: nicer!
				toRemove = i
				break
			}

			joinParts := strings.SplitN(join, ".", 2)
			// Get the collection from meta defined by joinParts[0]
			localCollection, err := collectionMetaGetter(joinParts[0])
			if err != nil {
				return nil, err
			}

			// Make our joinCollection repr
			joinCollection := NewCollection(localCollection)

			// Now we add our reference to all the things that we relate to
			joinField := localCollection.GetFieldByName(joinParts[1])
			reverseJoin := &ReverseJoin{
				Key:       joinParts[1],
				C:         joinCollection,
				JoinField: joinField,
				Filter:    joinMap[join],
			}
			// TODO: we *could* support it, but it would be odd, because we
			// overwrite the joinkey later (since it has to match to join)
			// we can change this in the future to only do the lookup if the
			// key matches? (since thats presumably what people want?)
			if _, ok := reverseJoin.Filter[reverseJoin.Key]; ok {
				return nil, fmt.Errorf("Join filter cannot include the join field")
			}

			foreignCollectionList, ok := collectionMap[joinField.Relation.Collection]
			// If we don't have a way to join to the designated remote, then skip
			if !ok {
				continue
			}
			for _, foreignCollection := range foreignCollectionList {
				foreignCollection.ReverseJoin = append(foreignCollection.ReverseJoin, reverseJoin)
			}

			// Now that we've joined in a new table, lets add it to our map
			collectionList, ok := collectionMap[reverseJoin.C.Name]
			if !ok {
				collectionList = []*Collection{}
			}
			collectionMap[reverseJoin.C.Name] = append(collectionList, reverseJoin.C)

			// Now we add ourselves as a top-level collection (in case others need to join from us)
			collectionMapItems, ok := collectionMap[localCollection.GetName()]
			if !ok {
				collectionMapItems = make([]*Collection, 0, 1)
			}

			joinRadixTree.Insert(join, joinCollection)

			collectionMapItems = append(collectionMapItems, joinCollection)
			toRemove = i
			break
		}

		if toRemove >= 0 {
			reverseJoinList = append(reverseJoinList[:toRemove], reverseJoinList[toRemove+1:]...)
		}

	}

	if len(forwardJoinMap) > 0 {
		for k, v := range forwardJoinMap {
			if err := forwardJoinFunc(k, v); err != nil {
				return nil, err
			}
		}
	}

	return thisCollection, nil
}
