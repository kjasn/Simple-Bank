package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kjasn/simple-bank/api"
	db "github.com/kjasn/simple-bank/db/sqlc"
	_ "github.com/kjasn/simple-bank/doc/statik"
	"github.com/kjasn/simple-bank/gapi"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/utils"
	_ "github.com/lib/pq" // provide a driver that implements postgres
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func runDBMigration(migrationURL string, postgresDNS string) {
	migration, err := migrate.New(migrationURL, postgresDNS)
	if err != nil {
		log.Fatal("cannot create new migrate instance: ", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange { // ignore no change
		log.Fatal("cannot run migration up: ", err)
	}

	log.Println("run migrate successfully")
}

func main() {
	config, err := utils.LoadConfig(".")

	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DSN)
	if err != nil {
		log.Fatal("Fail to connect to the db:", err)
	}

	// 
	runDBMigration(config.MigrationURL, config.DSN)
	store := db.NewStore(conn)

	go runGatewayServer(config, store)
	runGrpcServer(config, store)

}

func runGrpcServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot not create gRPC server: ", err)
	}

	// create grpc server instance
	grpcServer := grpc.NewServer()

	// register custom server into grpc server
	pb.RegisterSimpleBankServer(grpcServer, server)

	// create a reflection
	reflection.Register(grpcServer)

	// listen and start
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot not create listener: ", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot not start gRPC server: ", err)
	}
}

func runGatewayServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot not create gRPC server: ", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot not create handler: ", err)
	}

	// create HTTP server mux, receive HTTP requests from clients
	mux := http.NewServeMux()
	// rerouter to gRPC mux
	mux.Handle("/", grpcMux)

	// create statik file server
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot not create statik file server: ", err)
	}
	// fs := http.FileServer(http.Dir("./doc/swagger"))
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	// listen and start
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot not create listener: ", err)
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot not start HTTP gateway server: ", err)
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