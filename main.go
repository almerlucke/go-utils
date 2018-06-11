package main

import (
	"log"

	"github.com/almerlucke/go-utils/sql"
)

type Test struct {
	sql.MySQLTable
	sql.Model
	NoField string `mysql:"-"`
	Name    string
	Count   int64  `mysql:"DEFAULT 2"`
	Other   string `mysql:"override,VARCHAR(12)"`
	When    sql.DateTime
	Blub    []byte
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
