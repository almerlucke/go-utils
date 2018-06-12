package main

import (
	"log"

	"github.com/almerlucke/go-utils/sql"
)

type Test struct {
	sql.MySQLTable
	sql.Model
	NoField string       `db:"-"`
	Name    string       `db:"name"`
	Count   int64        `db:"count" mysql:"DEFAULT 2"`
	Other   string       `db:"other" mysql:"override,VARCHAR(12)"`
	When    sql.DateTime `db:"when"`
	Blub    []byte       `db:"blub"`
}

func (test *Test) TableDescriptor() (*sql.MySQLTableDescriptor, error) {
	return sql.StructToMySQLTableDescriptor(test)
}

func (test *Test) TableName() string {
	return "test"
}

func (test *Test) TableKeysAndIndices() []string {
	return []string{
		"KEY `test` (`test`)",
	}
}

func (test *Test) TableQuery() (string, error) {
	return sql.TablerToMySQLStatement(test)
}

func main() {
	test := Test{}

	query, err := test.TableQuery()
	if err != nil {
		log.Fatalf("err %v", err)
	}

	log.Printf("%v", query)
}
