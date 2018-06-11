package sql

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/almerlucke/go-utils/reflection/structural"
)

// MySQLType MySQL column type
type MySQLType int

const (
	// MySQLUnknown unknown column type
	MySQLUnknown MySQLType = iota
	// MySQLTinyInt int 8 bits
	MySQLTinyInt
	// MySQLSmallInt int 16 bits
	MySQLSmallInt
	// MySQLInt int 32/64 bits depending on strconv.IntSize
	MySQLInt
	// MySQLBigInt int 64 bits
	MySQLBigInt
	// MySQLUnsignedTinyInt unsigned int 8 bits
	MySQLUnsignedTinyInt
	// MySQLUnsignedSmallInt unsigned int 16 bits
	MySQLUnsignedSmallInt
	// MySQLUnsignedInt unsigned int 32/64 bits depending on strconv.IntSize
	MySQLUnsignedInt
	// MySQLUnsignedBigInt unsigned int 64 bits
	MySQLUnsignedBigInt
	// MySQLFloat float 32
	MySQLFloat
	// MySQLDouble float 64
	MySQLDouble
	// MySQLBool bool -> tinyint(1)
	MySQLBool
	// MySQLText string text
	MySQLText
	// MySQLBlob []byte
	MySQLBlob
	// MySQLDateTime DateTime
	MySQLDateTime
	// MySQLDate Date
	MySQLDate
)

// MySQLTabler interface for structs that represent a MySQL table
type MySQLTabler interface {
	TableEngine() string
	TableCharSet() string
	TableName() string
	TableKeysAndIndices() []string
	TableDescriptor() (*MySQLTableDescriptor, error)
	TableQuery() (string, error)
}

// MySQLTable is a default table to embed in a real table struct
// it defines the default engine and character set
type MySQLTable struct{}

// TableEngine returns the default MySQL table engine InnoDB
func (table *MySQLTable) TableEngine() string {
	return "InnoDB"
}

// TableCharSet returns the default MySQL table charset utf8mb4
func (table *MySQLTable) TableCharSet() string {
	return "utf8mb4"
}

// TableKeysAndIndices default returns empty slice
func (table *MySQLTable) TableKeysAndIndices() []string {
	return []string{}
}

// MySQLColumnDescriptor column descriptor, is used by StructToMySQLTableDescriptor
// to store column info from struct field and tags
type MySQLColumnDescriptor struct {
	Name         string
	Type         MySQLType
	Raw          string
	OverrideType bool
	IsPrimary    bool
}

// MySQLTableDescriptor table descriptor, is used by StructToMySQLTableDescriptor
// to store table column info
type MySQLTableDescriptor struct {
	PrimaryColumn *MySQLColumnDescriptor
	Columns       []*MySQLColumnDescriptor
}

// String returns column descriptor MySQL query string
func (column *MySQLColumnDescriptor) String() string {
	if column.OverrideType {
		return fmt.Sprintf("`%v` %v", column.Name, column.Raw)
	}

	if column.Raw == "" {
		return fmt.Sprintf("`%v` %v", column.Name, column.Type.String())
	}

	return fmt.Sprintf("`%v` %v %v", column.Name, column.Type.String(), column.Raw)
}

// String returns type as MySQL query string
func (t MySQLType) String() string {
	switch t {
	case MySQLTinyInt:
		return "tinyint"
	case MySQLSmallInt:
		return "smallint"
	case MySQLInt:
		return "int"
	case MySQLBigInt:
		return "bigint"
	case MySQLUnsignedTinyInt:
		return "tinyint unsigned"
	case MySQLUnsignedSmallInt:
		return "smallint unsigned"
	case MySQLUnsignedInt:
		return "int unsigned"
	case MySQLUnsignedBigInt:
		return "bigint unsigned"
	case MySQLFloat:
		return "float"
	case MySQLDouble:
		return "double"
	case MySQLBool:
		return "tinyint(1)"
	case MySQLText:
		return "text"
	case MySQLBlob:
		return "blob"
	case MySQLDate:
		return "date"
	case MySQLDateTime:
		return "datetime"
	default:
		return ""
	}
}

// Model can be used as basis for records that can be updated and deleted
type Model struct {
	ID         uint64   `json:"id" db:"id" mysql:"NOT NULL AUTO_INCREMENT"`
	CreatedAt  DateTime `json:"createdAt" db:"created_at" mysql:"DEFAULT CURRENT_TIMESTAMP"`
	ModifiedAt DateTime `json:"modifiedAt" db:"modified_at" mysql:"DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Deleted    bool     `json:"-" db:"deleted" mysql:"DEFAULT 0"`
}

func parseMySQLTag(tag string, columnDesc *MySQLColumnDescriptor) bool {
	skipColumn := false
	components := strings.Split(tag, ",")

	for _, component := range components {
		if component == "-" {
			skipColumn = true
		} else if component == "override" {
			columnDesc.OverrideType = true
		} else if component == "primary" {
			columnDesc.IsPrimary = true
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

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func nameToMySQLName(name string) string {
	snake := matchFirstCap.ReplaceAllString(name, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func fieldToMySQLType(field structural.FieldDescriptor) MySQLType {
	t := field.Type()
	kind := t.Kind()

	switch kind {
	case reflect.Int:
		if strconv.IntSize == 32 {
			return MySQLInt
		} else if strconv.IntSize == 64 {
			return MySQLBigInt
		}
	case reflect.Int8:
		return MySQLTinyInt
	case reflect.Int16:
		return MySQLSmallInt
	case reflect.Int32:
		return MySQLInt
	case reflect.Int64:
		return MySQLBigInt
	case reflect.Uint:
		if strconv.IntSize == 32 {
			return MySQLUnsignedInt
		} else if strconv.IntSize == 64 {
			return MySQLUnsignedBigInt
		}
	case reflect.Uint8:
		return MySQLUnsignedTinyInt
	case reflect.Uint16:
		return MySQLUnsignedSmallInt
	case reflect.Uint32:
		return MySQLUnsignedInt
	case reflect.Uint64:
		return MySQLUnsignedBigInt
	case reflect.Float32:
		return MySQLFloat
	case reflect.Float64:
		return MySQLDouble
	case reflect.String:
		return MySQLText
	case reflect.Bool:
		return MySQLBool
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return MySQLBlob
		}
	default:
		if field.Type().PkgPath() == "github.com/almerlucke/go-utils/sql" {
			typeName := field.Type().Name()
			if typeName == "Date" {
				return MySQLDate
			} else if typeName == "DateTime" {
				return MySQLDateTime
			}
		}
	}

	return MySQLUnknown
}

// StructToMySQLTableDescriptor generates column and table info from structure fields and mysql tags
func StructToMySQLTableDescriptor(obj interface{}) (*MySQLTableDescriptor, error) {
	desc, ok := structural.NewStructDescriptor(obj)
	if !ok {
		return nil, fmt.Errorf("can't get struct descriptor from object %v", obj)
	}

	tableDesc := &MySQLTableDescriptor{
		Columns: []*MySQLColumnDescriptor{},
	}

	var primaryColumn *MySQLColumnDescriptor

	err := desc.ScanFields(true, true, nil, func(field structural.FieldDescriptor, context interface{}) error {
		if field.Anonymous() {
			return nil
		}

		fieldTag := field.Tag().Get("mysql")
		fieldName := field.Name()

		columnDesc := &MySQLColumnDescriptor{
			Type: fieldToMySQLType(field),
			Name: nameToMySQLName(fieldName),
		}

		skipColumn := false

		if fieldTag != "" {
			skipColumn = parseMySQLTag(fieldTag, columnDesc)
		}

		if !skipColumn {
			if columnDesc.Type == MySQLUnknown && !columnDesc.OverrideType {
				return fmt.Errorf("unmappable field %v", field)
			}

			if columnDesc.IsPrimary {
				primaryColumn = columnDesc
			}

			tableDesc.Columns = append(tableDesc.Columns, columnDesc)
		}

		return nil
	})

	if primaryColumn == nil && len(tableDesc.Columns) > 0 {
		tableDesc.PrimaryColumn = tableDesc.Columns[0]
	}

	return tableDesc, err
}

// TablerToMySQLStatement returns a MySQL create table query from a MySQLTabler object
func TablerToMySQLStatement(tabler MySQLTabler) (string, error) {
	desc, err := tabler.TableDescriptor()
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("CREATE TABLE `%v` (\n", tabler.TableName()))

	entries := []string{}
	for _, column := range desc.Columns {
		entries = append(entries, column.String())
	}

	if desc.PrimaryColumn != nil {
		entries = append(entries, fmt.Sprintf("PRIMARY KEY (`%v`)", desc.PrimaryColumn.Name))
	}

	for _, key := range tabler.TableKeysAndIndices() {
		entries = append(entries, key)
	}

	endIndex := len(entries) - 1
	for index, entry := range entries {
		if index != endIndex {
			buffer.WriteString(fmt.Sprintf("\t%v,\n", entry))
		} else {
			buffer.WriteString(fmt.Sprintf("\t%v\n", entry))
		}
	}

	buffer.WriteString(fmt.Sprintf(") ENGINE=%v DEFAULT CHARSET=%v;", tabler.TableEngine(), tabler.TableCharSet()))

	return buffer.String(), nil
}
