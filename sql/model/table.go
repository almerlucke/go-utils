package model

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"reflect"

	sqlUtils "github.com/almerlucke/go-utils/sql"
)

// Tabler interface for structs that represent a MySQL table
type Tabler interface {
	TableEngine() string
	TableCharSet() string
	TableName() string
	TableKeysAndIndices() []string
	TableDescriptor() *TableDescriptor
	TableQuery() string
	Insert([]interface{}, sqlUtils.Queryer) (sql.Result, error)
	Select(string) *Select
	Update(interface{}, sqlUtils.Queryer) (sql.Result, error)
}

// Table is a definition of a SQL table and conforms to tabler interface
type Table struct {
	Engine         string
	CharSet        string
	Name           string
	KeysAndIndices []string
	Descriptor     *TableDescriptor
}

// NewTable creates a new table definition from a struct template
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
func (table *Table) Insert(objs []interface{}, queryer sqlUtils.Queryer) (sql.Result, error) {
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

// Select creates a select statement with From set to the table
func (table *Table) Select(fields string) *Select {
	return &Select{
		Fields: replaceStructFieldsWithSQLFields(fields, table.TemplateMap()),
		From:   table,
	}
}

// Update object, use primary key for where clause
func (table *Table) Update(obj interface{}, queryer sqlUtils.Queryer) (sql.Result, error) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("UPDATE %v SET ", table.Name))

	desc := table.Descriptor
	values := []interface{}{}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	addComma := false

	// Add column names to update query
	for _, column := range desc.Columns {
		if column == desc.PrimaryColumn {
			continue
		}

		if addComma {
			buffer.WriteRune(',')
		} else {
			addComma = true
		}

		buffer.WriteString(fmt.Sprintf("`%v`=?", column.Name))

		// Get field value
		f := v.FieldByName(column.ActualName)
		values = append(values, f.Interface())
	}

	buffer.WriteString(fmt.Sprintf(" WHERE `%v`=?", desc.PrimaryColumn.Name))

	f := v.FieldByName(desc.PrimaryColumn.ActualName)
	values = append(values, f.Interface())

	log.Printf("update query: %v\n", buffer.String())

	return queryer.Exec(buffer.String(), values...)
}

// ResultType returns the reflect Type for the raw table structure
func (table *Table) ResultType() reflect.Type {
	return table.Descriptor.RawDescriptor.Type()
}

// FromStatement for Selectable interface
func (table *Table) FromStatement() string {
	return "`" + table.Name + "`"
}

// TemplateMap for Selectable interface
func (table *Table) TemplateMap() map[string]string {
	desc := table.Descriptor
	templateMap := map[string]string{}

	for k, v := range desc.ColumnMap {
		templateMap[k] = v.Name
	}

	return templateMap
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

// NewDBWithTables creates a new DB object initialized with tables
func NewDBWithTables(config *sqlUtils.Configuration, tables ...Tabler) (*sqlUtils.DB, error) {
	db, err := sqlUtils.New(config)
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
