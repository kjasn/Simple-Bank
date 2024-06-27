package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kjasn/simple-bank/utils"
)


var testStore Store

func TestMain(m *testing.M) {
	var err error

	config, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DSN)

	if err != nil {
		log.Fatal("Fail to connect to the db:", err)
	}

	testStore = NewStore(connPool)
	os.Exit(m.Run())
}