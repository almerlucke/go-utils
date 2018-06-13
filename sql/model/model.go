// Package model can be used to auto generate sql table and column information from struct with the db and sql tags.
// We use reflection to look at the name and type of fields combined with field tags to generate a table model.
// The "db" tag that is used by the sql package is also considered when getting field names. In this package a Tabler interface is
// defined which can also be used to insert and select. This is not a full fledged select implementation but can be used
// for quick access. You can still use raw queries like normal
package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/almerlucke/go-utils/reflection/structural"
	sqlUtils "github.com/almerlucke/go-utils/sql"
)

// Model can be used as basis for records that can be updated and deleted
type Model struct {
	ID         uint64            `json:"id" db:"id" sql:"auto,NOT NULL AUTO_INCREMENT"`
	CreatedAt  sqlUtils.DateTime `json:"createdAt" db:"created_at" sql:"auto,DEFAULT CURRENT_TIMESTAMP"`
	ModifiedAt sqlUtils.DateTime `json:"modifiedAt" db:"modified_at" sql:"auto,DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Deleted    bool              `json:"-" db:"deleted" sql:"auto,DEFAULT 0"`
}

// ColumnDescriptor column descriptor, is used by StructToTableDescriptor
// to store column info from struct field and tags
type ColumnDescriptor struct {
	Name         string
	Type         string
	Raw          string
	OverrideType bool
	IsPrimary    bool
	ActualName   string
	Auto         bool
}

// TableDescriptor table descriptor, is used by StructToTableDescriptor
// to store table column info
type TableDescriptor struct {
	RawDescriptor structural.StructDescriptor
	PrimaryColumn *ColumnDescriptor
	Columns       []*ColumnDescriptor
	ColumnMap     map[string]*ColumnDescriptor
}

// String returns column descriptor MySQL query string
func (column *ColumnDescriptor) String() string {
	if column.OverrideType {
		return fmt.Sprintf("`%v` %v", column.Name, column.Raw)
	}

	if column.Raw == "" {
		return fmt.Sprintf("`%v` %v", column.Name, column.Type)
	}

	return fmt.Sprintf("`%v` %v %v", column.Name, column.Type, column.Raw)
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func nameToMySQLName(name string) string {
	snake := matchFirstCap.ReplaceAllString(name, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func fieldToMySQLType(field structural.FieldDescriptor) string {
	t := field.Type()
	kind := t.Kind()

	switch kind {
	case reflect.Int:
		if strconv.IntSize == 32 {
			return "int"
		} else if strconv.IntSize == 64 {
			return "bigint"
		}
	case reflect.Int8:
		return "tinyint"
	case reflect.Int16:
		return "smallint"
	case reflect.Int32:
		return "int"
	case reflect.Int64:
		return "bigint"
	case reflect.Uint:
		if strconv.IntSize == 32 {
			return "int unsigned"
		} else if strconv.IntSize == 64 {
			return "bigint unsigned"
		}
	case reflect.Uint8:
		return "tinyint unsigned"
	case reflect.Uint16:
		return "smallint unsigned"
	case reflect.Uint32:
		return "int unsigned"
	case reflect.Uint64:
		return "bigint unsigned"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.String:
		return "text"
	case reflect.Bool:
		return "tinyint(1)"
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "blob"
		}
	default:
		if field.Type().PkgPath() == "github.com/almerlucke/go-utils/sql" {
			typeName := field.Type().Name()
			if typeName == "Date" {
				return "date"
			} else if typeName == "DateTime" {
				return "datetime"
			}
		}
	}

	return ""
}

func parseSQLTag(tag string, columnDesc *ColumnDescriptor) bool {
	skipColumn := false
	components := strings.Split(tag, ",")

	for _, component := range components {
		if component == "-" {
			skipColumn = true
		} else if component == "override" {
			columnDesc.OverrideType = true
		} else if component == "primary" {
			columnDesc.IsPrimary = true
		} else if component == "auto" {
			columnDesc.Auto = true
		} else if component != "" {
			defs := strings.SplitN(component, "=", 2)
			if len(defs) == 2 {
				if defs[0] == "name" {
					columnDesc.Name = defs[1]
				}
			} else {
				columnDesc.Raw = defs[0]
			}
		}
	}

	return skipColumn
}

// StructToTableDescriptor generates column and table info from structure fields and db/sql tags.
// The sql tag is a comma separated list of definitions. The following keywords are defined.
// - override: this indicates that the derived sql type should be replaced by the raw statement in the
//   sql tag
// - primary: this indicates that the fields is the primary key, otherwise the first field of the struct
//   will be taken as primary key
// - auto: this indicates that the field value will be auto generated, so is not included in the
//	 Insert method query result
// - name=name: can be used to override the derived name from "db" tag or field name
// - in all other cases the value is inserted as raw sql for a column in the CREATE table query
func StructToTableDescriptor(obj interface{}) (*TableDescriptor, error) {
	desc, ok := structural.NewStructDescriptor(obj)
	if !ok {
		return nil, fmt.Errorf("can't get struct descriptor from object %v", obj)
	}

	tableDesc := &TableDescriptor{
		RawDescriptor: desc,
		Columns:       []*ColumnDescriptor{},
		ColumnMap:     map[string]*ColumnDescriptor{},
	}

	var primaryColumn *ColumnDescriptor

	err := desc.ScanFields(true, true, nil, func(field structural.FieldDescriptor, context interface{}) error {
		if field.Anonymous() {
			return nil
		}

		fieldTag1 := field.Tag().Get("db")
		fieldTag2 := field.Tag().Get("sql")
		fieldName := field.Name()

		columnDesc := &ColumnDescriptor{
			Type:       fieldToMySQLType(field),
			Name:       nameToMySQLName(fieldName),
			ActualName: fieldName,
		}

		skipColumn := false

		if fieldTag1 != "" {
			if fieldTag1 == "-" {
				skipColumn = true
			} else {
				columnDesc.Name = fieldTag1
			}
		}

		if fieldTag2 != "" {
			skipColumn = skipColumn || parseSQLTag(fieldTag2, columnDesc)
		}

		if !skipColumn {
			if columnDesc.Type == "" && !columnDesc.OverrideType {
				return fmt.Errorf("unmappable field %v", field)
			}

			if columnDesc.IsPrimary {
				primaryColumn = columnDesc
			}

			tableDesc.Columns = append(tableDesc.Columns, columnDesc)
			tableDesc.ColumnMap[columnDesc.ActualName] = columnDesc
		}

		return nil
	})

	if primaryColumn == nil && len(tableDesc.Columns) > 0 {
		tableDesc.PrimaryColumn = tableDesc.Columns[0]
	}

	return tableDesc, err
}
