package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
	"github.com/kjasn/simple-bank/worker"
	_ "github.com/lib/pq" // provide a driver that implements postgres
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func runDBMigration(migrationURL string, postgresDNS string) {
	migration, err := migrate.New(migrationURL, postgresDNS)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange { // ignore no change
		log.Fatal().Err(err).Msg("Cannot run migration up")
	}

	log.Info().Msg("Run migrate successfully")
}

func main() {
	config, err := utils.LoadConfig(".")

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load config")
	}

	if config.Environment == "development" {
		// set good read style for console when development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})	
	}
	conn, err := sql.Open(config.DBDriver, config.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to connect to the db")
	}

	// 
	runDBMigration(config.MigrationURL, config.DSN)
	store := db.NewStore(conn)

	// connect to redis and create a distributor
	redisOpt := asynq.RedisClientOpt{Addr: config.RedisAddress}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	go runTaskProcessor(redisOpt, store)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)

}


func runTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store)

	log.Info().Msg("start task processor...")
	if err := taskProcessor.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runGrpcServer(config utils.Config, store db.Store, distributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, distributor)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create gRPC server")
	}

	// add logger for grpc request
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	// create grpc server instance
	grpcServer := grpc.NewServer(grpcLogger)

	// register custom server into grpc server
	pb.RegisterSimpleBankServer(grpcServer, server)

	// create a reflection
	reflection.Register(grpcServer)

	// listen and start
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create listener")
		// log.Fatal().Msgf("Cannot create listener: %s", err.Error())	
	}

	log.Info().Msgf("Start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start gRPC server")
	}
}

func runGatewayServer(config utils.Config, store db.Store, distributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, distributor)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create gRPC server")
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
		log.Fatal().Err(err).Msg("Cannot create handler")
	}

	// create HTTP server mux, receive HTTP requests from clients
	mux := http.NewServeMux()
	// rerouter to gRPC mux
	mux.Handle("/", grpcMux)

	// create statik file server
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create statik file server")
	}
	// fs := http.FileServer(http.Dir("./doc/swagger"))
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	// listen and start
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create listener")
	}

	log.Info().Msgf("start HTTP gateway server at %s", listener.Addr().String())
	// add logger handler
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start HTTP gateway server")
	}
}

func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create HTTP server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot start HTTP server")
	}
}