package main

import (
	"database/sql"
	"log"

	"github.com/kjasn/simple-bank/api"
	db "github.com/kjasn/simple-bank/db/sqlc"
	_ "github.com/lib/pq" // provide a driver that implements postgres
)

const (
	dbDriver = "postgres"	// can not directly use postgres as driver pass to sql.Open
	dsn = "postgres://root:8520@localhost:5432/simple_bank?sslmode=disable"
	address = "127.0.0.1:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal("Fail to connect to the db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(address)
	if err != nil {
		log.Fatal("Cannot not start server: ", err)
	}
}