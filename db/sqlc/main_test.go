package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/kjasn/simple-bank/utils"
	_ "github.com/lib/pq" // provide a driver that implements postgres
)


var testQueries *Queries	// save as a global var
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	config, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DSN)

	if err != nil {
		log.Fatal("Fail to connect to the db:", err)
	}

	testQueries = New(testDB)
	os.Exit(m.Run())
}