package metadata

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

/*

	Since field_types need to be used regardless of application, we are going to
	have a "FieldTypeRegistry" which is a central place to store all the field_types
	This will allow other us to inject all the other types (custom types) into the same subsystem


*/

type FieldTypeRegister struct {
	r map[string]*FieldType
	l *sync.RWMutex
}

func (r *FieldTypeRegister) Add(f *FieldType) error {
	r.l.Lock()
	defer r.l.Unlock()
	if strings.HasPrefix(f.Name, InternalFieldPrefix) {
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

type FieldType struct {
	Name        string                `json:"name"`
	DatamanType DatamanType           `json:"dataman_type"`
	Constraints []*ConstraintInstance `json:"constraints,omitempty"`
}

// Validate and normalize
func (f *FieldType) Normalize(val interface{}) (interface{}, error) {
	normalizedVal, err := f.DatamanType.Normalize(val)
	if err != nil {
		return normalizedVal, err
	}

	if f.Constraints != nil {
		for i, constraint := range f.Constraints {
			if !constraint.Func(normalizedVal) {
				return normalizedVal, fmt.Errorf("Failed constraint %d: %v", i, constraint)
			}
		}
	}

	return normalizedVal, nil
}

func (f *FieldType) Equal(o *FieldType) bool {
	// TODO: also compare constraints
	return f.Name == o.Name && f.DatamanType == o.DatamanType
}
