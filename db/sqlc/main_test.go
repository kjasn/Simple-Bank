package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // provide a driver that implements postgres
)

const (
	dbDriver = "postgres"	// can not directly use postgres as driver pass to sql.Open
	dsn = "postgres://root:8520@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries	// save as a global var

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dsn)

	if err != nil {
		log.Fatal("Fail to connect to the db:", err)
	}

	testQueries = New(conn)
	os.Exit(m.Run())
}