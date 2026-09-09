package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbDriver = "mysql"
	dbSource = "panyu:panyu@tcp(localhost:3306)/cloud-disk?multiStatements=true&parseTime=true"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("can not connect to %v", dbSource)
	}
	testDB = db
	testQueries = New(testDB)
	os.Exit(m.Run())
}
