// Package structural can be used to get reflection information about a given
// structure object. The object needs to be a structure or pointer to structure.
// ScanFields can be used to loop through all structure fields.
package structural

import (
	"errors"
	"fmt"
	"reflect"
)

/*
 *
 *  Other types
 *
 */

// ScanFunction defines a function that can be given to ScanFields method of
// StructDescriptor objects. The scan function receives a field descriptor for
// each field
type ScanFunction func(field FieldDescriptor, context interface{}) error

/*
 *
 *  Interfaces
 *
 */

// Descriptor combines utility methods around a structure or field
type Descriptor interface {

	// Type get type of described object
	Type() reflect.Type

	// Value get value of described object
	Value() reflect.Value

	// Kind get kind of described object
	Kind() reflect.Kind

	// UID unique id of described object, uniqueness is defined inside
	// a given set
	UID() string

	// Name get name of described object
	Name() string

	// PkgPath get package path of described object
	PkgPath() string

	// IsExportable check if described object is exported
	IsExportable() bool

	// CanSet true if described object can be set
	CanSet() bool
}

// FieldDescriptor describes a field
type FieldDescriptor interface {
	// Inherits from Descriptor interface
	Descriptor

	// Field underlying struct field
	Field() reflect.StructField

	// Anonymous true if field is embedded
	Anonymous() bool

	// Tag the tag assigned to this field
	Tag() reflect.StructTag

	// StructDescriptor create a StructDescriptor, returns an error when
	// field is not of struct or struct ptr kind
	StructDescriptor() (StructDescriptor, error)
}

// StructDescriptor describes a struct
type StructDescriptor interface {
	// Inherits from Descriptor interface
	Descriptor

	// ScanFields scan fields of struct descriptor with a scan function, the
	// scan function is passed a field descriptor and a user defined context.
	// If exportable is true, only fields that are exported are considered. If
	// embedded is true, embedded structs or struct ptr fields are scanned as
	// well.
	ScanFields(exportable bool, embedded bool, context interface{}, scanFunction ScanFunction) error

	// FieldByName get field descriptor by name
	FieldByName(name string) (FieldDescriptor, bool)
}

/*
 *
 *  Field descriptor implementation
 *
 */

type fieldDescriptorImp struct {
	reflect.StructField
	V reflect.Value
}

func (desc *fieldDescriptorImp) Type() reflect.Type {
	return desc.StructField.Type
}

func (desc *fieldDescriptorImp) Value() reflect.Value {
	return desc.V
}

func (desc *fieldDescriptorImp) Kind() reflect.Kind {
	return desc.StructField.Type.Kind()
}

func (desc *fieldDescriptorImp) Name() string {
	return desc.StructField.Name
}

func (desc *fieldDescriptorImp) PkgPath() string {
	return desc.StructField.PkgPath
}

func (desc *fieldDescriptorImp) UID() string {
	return fmt.Sprintf("%v.%v", desc.StructField.PkgPath, desc.StructField.Name)
}

func (desc *fieldDescriptorImp) IsExportable() bool {
	return desc.StructField.PkgPath == ""
}

func (desc *fieldDescriptorImp) Tag() reflect.StructTag {
	return desc.StructField.Tag
}

func (desc *fieldDescriptorImp) Anonymous() bool {
	return desc.StructField.Anonymous
}

func (desc *fieldDescriptorImp) CanSet() bool {
	return desc.V.CanSet()
}

func (desc *fieldDescriptorImp) Field() reflect.StructField {
	return desc.StructField
}

func (desc *fieldDescriptorImp) StructDescriptor() (StructDescriptor, error) {
	fieldType := desc.StructField.Type
	fieldKind := fieldType.Kind()

	// Check if kind is struct
	if fieldKind == reflect.Struct {
		return &structDescriptorImp{
			T: fieldType,
			V: desc.V,
		}, nil
	}

	// If ptr
	if fieldKind == reflect.Ptr {
		elem := fieldType.Elem()

		// Check if ptr to struct
		if elem.Kind() == reflect.Struct {
			return &structDescriptorImp{
				T: elem,
				V: desc.V.Elem(),
			}, nil
		}
	}

	return nil, errors.New("Field type is not a struct or struct ptr")
}

/*
 *
 *  Struct descriptor implementation
 *
 */

// Descriptor implementation struct
type structDescriptorImp struct {
	T reflect.Type
	V reflect.Value
}

func (desc *structDescriptorImp) Type() reflect.Type {
	return desc.T
}

func (desc *structDescriptorImp) Value() reflect.Value {
	return desc.V
}

func (desc *structDescriptorImp) Kind() reflect.Kind {
	return desc.T.Kind()
}

func (desc *structDescriptorImp) Name() string {
	return desc.T.Name()
}

func (desc *structDescriptorImp) PkgPath() string {
	return desc.T.PkgPath()
}

func (desc *structDescriptorImp) UID() string {
	return fmt.Sprintf("%v.%v", desc.T.PkgPath(), desc.T.Name())
}

func (desc *structDescriptorImp) IsExportable() bool {
	return desc.T.PkgPath() == ""
}

func (desc *structDescriptorImp) CanSet() bool {
	return desc.V.CanSet()
}

func (desc *structDescriptorImp) FieldByName(name string) (FieldDescriptor, bool) {
	field, ok := desc.T.FieldByName(name)
	if !ok {
		return nil, false
	}

	value := desc.V.FieldByName(name)

	return &fieldDescriptorImp{
		StructField: field,
		V:           value,
	}, true
}

func (desc *structDescriptorImp) ScanFields(exportable bool, embedded bool, context interface{}, scanFunction ScanFunction) error {
	// Get number of fields
	numField := desc.T.NumField()

	// Loop through fields
	for i := 0; i < numField; i++ {
		fieldDesc := &fieldDescriptorImp{
			StructField: desc.T.Field(i),
			V:           desc.V.Field(i),
		}

		// Check if we need to scan this field, if exportable is true and
		// the field is not exportable, we do not scan it
		if !exportable || (exportable && fieldDesc.IsExportable()) {
			// Check if we want to scan embedded fields and if the field is
			// actually embedded
			if embedded && fieldDesc.Anonymous() {
				// Create embedded structure descriptor
				edesc, err := fieldDesc.StructDescriptor()
				if err != nil {
					return err
				}

				// Scan embedded descriptor fields
				err = edesc.ScanFields(exportable, embedded, context, scanFunction)
				if err != nil {
					return err
				}
			} else {
				// Pass field descriptor to the scan function
				err := scanFunction(fieldDesc, context)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

/*
 *
 *  Non interface functions
 *
 */

// NewStructDescriptor create struct descriptor from object, object can be struct or ptr to struct
func NewStructDescriptor(obj interface{}) (StructDescriptor, bool) {
	desc := &structDescriptorImp{}
	typeOf := reflect.TypeOf(obj)

	// Check for nil interface
	if typeOf == nil {
		return nil, false
	}

	// Get kind from type
	kind := typeOf.Kind()

	// Check if struct kind
	if kind == reflect.Struct {
		desc.T = typeOf
		desc.V = reflect.ValueOf(obj)
		return desc, true
	}

	// If kind is ptr
	if kind == reflect.Ptr {
		elem := typeOf.Elem()

		// Check if elem kind is struct
		if elem.Kind() == reflect.Struct {
			desc.T = elem
			desc.V = reflect.ValueOf(obj).Elem()
			return desc, true
		}
	}

	// Not a struct of ptr to struct
	return nil, false
}
