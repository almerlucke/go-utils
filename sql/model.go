package sql

import (
	"bytes"
	"database/sql"
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

// Tabler interface for structs that represent a MySQL table
type Tabler interface {
	TableEngine() string
	TableCharSet() string
	TableName() string
	TableKeysAndIndices() []string
	TableDescriptor() *TableDescriptor
	TableQuery() string
	Insert(objs []interface{}, queryer Queryer) (sql.Result, error)
	Select(fields ...string) *Select
}

// Table is a definition of a MySQL table and conforms to tabler interface
type Table struct {
	Engine         string
	CharSet        string
	Name           string
	KeysAndIndices []string
	Descriptor     *TableDescriptor
}

// NewTable creates a new table definition
func NewTable(name string, template interface{}) (*Table, error) {
	table := &Table{
		Engine:         "InnoDB",
		CharSet:        "utf8mb4",
		Name:           name,
		KeysAndIndices: []string{},
	}

	desc, err := StructToTableDescriptor(template)
	if err != nil {
		return nil, err
	}

	table.Descriptor = desc

	return table, nil
}

// TableEngine returns the table's engine
func (table *Table) TableEngine() string {
	return table.Engine
}

// TableCharSet returns the table's char set
func (table *Table) TableCharSet() string {
	return table.CharSet
}

// TableName returns the table's name
func (table *Table) TableName() string {
	return table.Name
}

// TableKeysAndIndices returns an array of raw KEY or INDEX definitions
func (table *Table) TableKeysAndIndices() []string {
	return table.KeysAndIndices
}

// TableDescriptor returns a descriptor of the table object
func (table *Table) TableDescriptor() *TableDescriptor {
	return table.Descriptor
}

// TableQuery returns a query string to CREATE the table
func (table *Table) TableQuery() string {
	return TablerToQuery(table)
}

// Insert objects into the table
func (table *Table) Insert(objs []interface{}, queryer Queryer) (sql.Result, error) {
	desc := table.Descriptor

	var buffer bytes.Buffer
	values := []interface{}{}

	buffer.WriteString(fmt.Sprintf("INSERT INTO `%v` (", table.Name))

	addComma := false
	numValues := 0

	for _, column := range desc.Columns {
		if column.Auto {
			continue
		} else {
			if addComma {
				buffer.WriteRune(',')
			} else {
				addComma = true
			}

			buffer.WriteString(column.Name)

			numValues++
		}
	}

	buffer.WriteString(") VALUES ")

	addComma = false
	for _, obj := range objs {

		if addComma {
			buffer.WriteRune(',')
		} else {
			addComma = true
		}

		t := reflect.TypeOf(obj)
		v := reflect.ValueOf(obj)
		if t.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		innerAddComma := false

		buffer.WriteRune('(')

		for _, column := range desc.Columns {
			if column.Auto {
				continue
			} else {
				if innerAddComma {
					buffer.WriteRune(',')
				} else {
					innerAddComma = true
				}

				buffer.WriteRune('?')

				values = append(values, v.FieldByName(column.ActualName).Interface())
			}
		}

		buffer.WriteRune(')')
	}

	return queryer.Exec(buffer.String(), values...)
}

// ColumnDescriptor column descriptor, is used by StructToTableDescriptor
// to store column info from struct field and tags
type ColumnDescriptor struct {
	Name         string
	Type         MySQLType
	Raw          string
	OverrideType bool
	IsPrimary    bool
	ActualName   string
	Auto         bool
}

// TableDescriptor table descriptor, is used by StructToTableDescriptor
// to store table column info
type TableDescriptor struct {
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
	ID         uint64   `json:"id" db:"id" mysql:"auto,NOT NULL AUTO_INCREMENT"`
	CreatedAt  DateTime `json:"createdAt" db:"created_at" mysql:"auto,DEFAULT CURRENT_TIMESTAMP"`
	ModifiedAt DateTime `json:"modifiedAt" db:"modified_at" mysql:"auto,DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Deleted    bool     `json:"-" db:"deleted" mysql:"auto,DEFAULT 0"`
}

func parseMySQLTag(tag string, columnDesc *ColumnDescriptor) bool {
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

// StructToTableDescriptor generates column and table info from structure fields and mysql tags
func StructToTableDescriptor(obj interface{}) (*TableDescriptor, error) {
	desc, ok := structural.NewStructDescriptor(obj)
	if !ok {
		return nil, fmt.Errorf("can't get struct descriptor from object %v", obj)
	}

	tableDesc := &TableDescriptor{
		Columns:   []*ColumnDescriptor{},
		ColumnMap: map[string]*ColumnDescriptor{},
	}

	var primaryColumn *ColumnDescriptor

	err := desc.ScanFields(true, true, nil, func(field structural.FieldDescriptor, context interface{}) error {
		if field.Anonymous() {
			return nil
		}

		fieldTag1 := field.Tag().Get("db")
		fieldTag2 := field.Tag().Get("mysql")
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
			skipColumn = skipColumn || parseMySQLTag(fieldTag2, columnDesc)
		}

		if !skipColumn {
			if columnDesc.Type == MySQLUnknown && !columnDesc.OverrideType {
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

// TablerToQuery returns a create table query from a Tabler object
func TablerToQuery(tabler Tabler) string {
	desc := tabler.TableDescriptor()

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%v` (\n", tabler.TableName()))

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

	return buffer.String()
}

// NewWithTables creates a new DB object initialized with tables
func NewWithTables(config *Configuration, tables ...Tabler) (*DB, error) {
	db, err := New(config)
	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		_, err = db.Exec(table.TableQuery())
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

/*



Tabler.Select("ID", "CreatedAt").Where("{{ID}} > 123 AND {{CreatedAt}} < NOW").GroupBy("").OrderBy("").Limit(0, 10)

*/

func (table *Table) Select(fields ...string) *Select {
	sqlFields := make([]string, len(fields))
	desc := table.Descriptor

	// Get sql names for struct field names
	for index, field := range fields {
		sqlFields[index] = desc.ColumnMap[field].Name
	}

	return &Select{
		Fields: sqlFields,
		From:   table,
	}
}

type Select struct {
	Fields            []string
	From              Tabler
	WhereCondition    string
	GroupByExpression string
	OrderByExpression string
	LimitResults      *Limit
}

func (sel *Select) replaceStructFieldsWithSqlFields(template string) string {
	r := regexp.MustCompile(`\{\{.+?\}\}`)
	desc := sel.From.TableDescriptor()

	return string(r.ReplaceAllFunc([]byte(template), func(src []byte) []byte {
		fieldName := strings.Trim(string(src), "{{}}")
		column := desc.ColumnMap[fieldName]
		if column != nil {
			return []byte(column.Name)
		}

		return []byte{}
	}))
}

func (sel *Select) Where(cond string) *Select {
	sel.WhereCondition = sel.replaceStructFieldsWithSqlFields(cond)
	return sel
}

func (sel *Select) GroupBy(cond string) *Select {
	sel.GroupByExpression = sel.replaceStructFieldsWithSqlFields(cond)
	return sel
}

func (sel *Select) OrderBy(expr string) *Select {
	sel.OrderByExpression = sel.replaceStructFieldsWithSqlFields(expr)
	return sel
}

func (sel *Select) Limit(offset int64, rowCount int64) *Select {
	sel.LimitResults = &Limit{
		Offset:   offset,
		RowCount: rowCount,
	}
	return sel
}

type Limit struct {
	Offset   int64
	RowCount int64
}
