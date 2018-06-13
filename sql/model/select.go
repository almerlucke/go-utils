package model

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	sqlUtils "github.com/almerlucke/go-utils/sql"
)

// Selectable can be used as From in Select setup
type Selectable interface {
	FromStatement() string
	TemplateMap() map[string]string
	ResultType() reflect.Type
}

// Select definition for creating select statements
type Select struct {
	Fields            string
	From              Selectable
	Alias             string
	WhereCondition    string
	GroupByExpression string
	OrderByExpression string
	LimitResults      *Limit
}

// NewSelect creates a new select statement
func NewSelect(fields string, from Selectable) *Select {
	return &Select{
		From:   from,
		Fields: replaceStructFieldsWithSQLFields(fields, from.TemplateMap()),
	}
}

// replaceStructFieldsWithSqlFields replaces handlebar template fields with structure field names
// for the real database fields
func replaceStructFieldsWithSQLFields(template string, templateMap map[string]string) string {
	r := regexp.MustCompile(`\{\{.+?\}\}`)

	return string(r.ReplaceAllFunc([]byte(template), func(src []byte) []byte {
		fieldName := strings.Trim(string(src), "{{}}")
		name := templateMap[fieldName]

		if name != "" {
			return []byte("`" + name + "`")
		}

		return []byte{}
	}))
}

// As adds an alias to the from statement
func (sel *Select) As(alias string) *Select {
	sel.Alias = replaceStructFieldsWithSQLFields(alias, sel.From.TemplateMap())
	return sel
}

// Where adds a where clause to the select definition
func (sel *Select) Where(cond string) *Select {
	sel.WhereCondition = replaceStructFieldsWithSQLFields(cond, sel.From.TemplateMap())
	return sel
}

// GroupBy adds a group by clause to the select definition
func (sel *Select) GroupBy(cond string) *Select {
	sel.GroupByExpression = replaceStructFieldsWithSQLFields(cond, sel.From.TemplateMap())
	return sel
}

// OrderBy adds a order by clause to the select definition
func (sel *Select) OrderBy(expr string) *Select {
	sel.OrderByExpression = replaceStructFieldsWithSQLFields(expr, sel.From.TemplateMap())
	return sel
}

// Limit adds a limit clause to the select definition
func (sel *Select) Limit(offset int64, rowCount int64) *Select {
	sel.LimitResults = &Limit{
		Offset:   offset,
		RowCount: rowCount,
	}
	return sel
}

// FromStatement for Selectable
func (sel *Select) FromStatement() string {
	return "(" + sel.Query() + ")"
}

// TemplateMap for Selectable
func (sel *Select) TemplateMap() map[string]string {
	// Pass back From template map
	return sel.From.TemplateMap()
}

// ResultType for Selectable
func (sel *Select) ResultType() reflect.Type {
	return sel.From.ResultType()
}

// Select for nested Select
func (sel *Select) Select(fields string) *Select {
	return &Select{
		From:   sel,
		Fields: replaceStructFieldsWithSQLFields(fields, sel.TemplateMap()),
	}
}

// Query string from Select object
func (sel *Select) Query() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("SELECT %v FROM %v", sel.Fields, sel.From.FromStatement()))

	if sel.Alias != "" {
		buffer.WriteString(fmt.Sprintf(" AS %v", sel.Alias))
	}

	if sel.WhereCondition != "" {
		buffer.WriteString(fmt.Sprintf(" WHERE %v", sel.WhereCondition))
	}

	if sel.GroupByExpression != "" {
		buffer.WriteString(fmt.Sprintf(" GROUP BY %v", sel.GroupByExpression))
	}

	if sel.OrderByExpression != "" {
		buffer.WriteString(fmt.Sprintf(" ORDER BY %v", sel.OrderByExpression))
	}

	if sel.LimitResults != nil {
		buffer.WriteString(fmt.Sprintf(" LIMIT %v, %v", sel.LimitResults.Offset, sel.LimitResults.RowCount))
	}

	return buffer.String()
}

// Run the select query
func (sel *Select) Run(queryer sqlUtils.Queryer, args ...interface{}) (interface{}, error) {
	resultType := sel.From.ResultType()
	v := reflect.New(reflect.SliceOf(reflect.PtrTo(resultType)))

	err := queryer.Select(v.Interface(), sel.Query(), args...)
	if err != nil {
		return nil, err
	}

	return v.Elem().Interface(), nil
}

// Limit offset and row count
type Limit struct {
	Offset   int64
	RowCount int64
}
