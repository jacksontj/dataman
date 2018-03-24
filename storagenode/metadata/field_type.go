package metadata

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/jacksontj/dataman/datamantype"
)

var InternalFieldTypePrefix = "_"

/*

	Since field_types need to be used regardless of application, we are going to
	have a "FieldTypeRegistry" which is a central place to store all the field_types
	This will allow other us to inject all the other types (custom types) into the same subsystem


*/

// TODO: exclude internal fields from serialization? Or have separate maps for internal vs user-defined?
type FieldTypeRegister struct {
	r map[string]*FieldType
	l *sync.RWMutex
}

func (r *FieldTypeRegister) Add(f *FieldType) error {
	r.l.Lock()
	defer r.l.Unlock()
	if strings.HasPrefix(f.Name, InternalFieldTypePrefix) {
		return fmt.Errorf("Reserved namespace!")
	}
	if _, ok := r.r[f.Name]; ok {
		return fmt.Errorf("Field type of that name already exists")
	}
	r.r[f.Name] = f
	return nil
}
func (r *FieldTypeRegister) Get(name string) *FieldType {
	r.l.RLock()
	defer r.l.RUnlock()
	return r.r[name]
}
func (r *FieldTypeRegister) Merge(o *FieldTypeRegister) {
	// TODO: if the `o` is being mutated this can be a problem. For now since this is just serialization stuff
	// I'm not bothering, if we go down that path we'll probably want to do some channel thing to avoid deadlocks
	r.l.Lock()
	defer r.l.Unlock()

	for name, fieldType := range o.r {
		if strings.HasPrefix(name, InternalFieldTypePrefix) {
			continue
		}
		if _, ok := r.r[name]; !ok {
			r.r[name] = fieldType
		}
	}
}

func (r *FieldTypeRegister) MarshalJSON() ([]byte, error) {
	r.l.RLock()
	defer r.l.RUnlock()
	return json.Marshal(r.r)
}

func (r *FieldTypeRegister) UnmarshalJSON(data []byte) error {
	r.r = make(map[string]*FieldType)
	if err := json.Unmarshal(data, &r.r); err != nil {
		return err
	}
	r.l = &sync.RWMutex{}

	return nil
}

// TODO: encapsulate in a struct (for locking etc.)
var FieldTypeRegistry *FieldTypeRegister

func init() {
	initFieldTypeRegistry()
}

func initFieldTypeRegistry() {
	if FieldTypeRegistry != nil {
		return
	}
	FieldTypeRegistry = &FieldTypeRegister{
		r: make(map[string]*FieldType),
		l: &sync.RWMutex{},
	}

	for _, fieldType := range listInternalFieldTypes() {
		FieldTypeRegistry.r[fieldType.Name] = fieldType
	}
}

// TODO: have this register the type? Right now this assumes this is in-sync with field_type_internal.go (which is bad to do)
func DatamanTypeToFieldType(f datamantype.DatamanType) *FieldType {
	return &FieldType{
		Name:        "_" + string(f),
		DatamanType: f,
	}
}

type FieldType struct {
	Name        string                  `json:"name"`
	DatamanType datamantype.DatamanType `json:"dataman_type"`
	Constraints []*ConstraintInstance   `json:"constraints,omitempty"`
}

// Validate and normalize
func (f *FieldType) Normalize(val interface{}) (interface{}, error) {
	normalizedVal, err := f.DatamanType.Normalize(val)
	if err != nil {
		return normalizedVal, err
	}

	if f.Constraints != nil {
		for _, constraint := range f.Constraints {
			if !constraint.Func(normalizedVal) {
				return normalizedVal, fmt.Errorf(constraint.ValidationError)
			}
		}
	}

	return normalizedVal, nil
}

func (f *FieldType) Equal(o *FieldType) bool {
	// TODO: also compare constraints
	return f.Name == o.Name && f.DatamanType == o.DatamanType
}
