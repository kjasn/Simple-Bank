package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/kjasn/simple-bank/api"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/gapi"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/utils"
	_ "github.com/lib/pq" // provide a driver that implements postgres
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	runGrpcServer(config, store)

}

func runGrpcServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot not create gRPC server: ", err)
	}

	grpcServer := grpc.NewServer()

	// register server
	pb.RegisterSimpleBankServer(grpcServer, server)

	// create a reflection
	reflection.Register(grpcServer)

	// listen and start
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot not create listener")
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot not start gRPC server")
	}
}

func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot not create server: ", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("Cannot not start server: ", err)
	}
}