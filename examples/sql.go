package examples

import (
	"log"

	"github.com/almerlucke/go-utils/sql/database"
	"github.com/almerlucke/go-utils/sql/migration"
	"github.com/almerlucke/go-utils/sql/model"
	"github.com/almerlucke/go-utils/sql/types"
	"github.com/almerlucke/go-utils/sql/utils"

	// Setup sql driver implementation, this needs to be done in main package
	_ "github.com/go-sql-driver/mysql"
)

// TestTable test table
type TestTable struct {
	model.Model
	StrTest     types.String   `db:"str_test" json:"strTest"`
	TimeTest    types.DateTime `db:"time_test" json:"timeTest"`
	DateTest    types.Date     `db:"date_test" json:"dateTest"`
	BoomSelecta int64          `db:"boom_selecta" json:"boomSelecta"`
	Tjekkeroo   int64          `db:"tjekkeroo" json:"tjekkeroo"`
}

// TestSQL test sql
func TestSQL(host string, user string, password string, dbName string) {
	testTable, err := model.NewTable("test", &TestTable{})
	if err != nil {
		log.Fatalf("err %v", err)
	}

	db, err := utils.NewDatabase(
		database.NewConfiguration(host, user, password, dbName),
		"1.0",
		[]*migration.Version{
			migration.NewVersion("1.0", []migration.Migration{
				migration.NewCustomMigration(func(queryer database.Queryer) error {
					log.Printf("migrate!")
					return nil
				}),
			}),
		},
		testTable,
	)

	if err != nil {
		log.Fatalf("err %v", err)
	}

	testTable.Insert([]interface{}{
		&TestTable{
			TimeTest: types.NewDateTime(),
			DateTest: types.NewDate(),
			StrTest:  "test1",
		},
		&TestTable{
			TimeTest: types.NewDateTime(),
			DateTest: types.NewDate(),
			StrTest:  "test2",
		},
	}, db)

	result, err := testTable.Select("*").Run(db)
	if err != nil {
		log.Fatalf("err %v", err)
	}

	for _, test := range result.([]*TestTable) {
		log.Printf("select %v", *test)
	}

	t := (result.([]*TestTable))[0]
	t.BoomSelecta = 64
	t.Tjekkeroo = 64
	t.StrTest = "excellent"

	testTable.Update(t, db)
}
