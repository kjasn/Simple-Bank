package main

import (
	"database/sql"
	"log"

	"github.com/kjasn/simple-bank/api"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
	_ "github.com/lib/pq" // provide a driver that implements postgres
)




func main() {
	config, err := utils.LoadConfig(".")

	if err != nil {
		log.Fatal("Cannot load config:", err)
	}


	conn, err := sql.Open(config.DBDriver, config.DSN)
	if err != nil {
		log.Fatal("Fail to connect to the db:", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot not create server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot not start server: ", err)
	}
}