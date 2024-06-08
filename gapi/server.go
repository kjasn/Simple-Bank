package gapi

import (
	"fmt"

	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/token"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/worker"
)

// Server servers gRPC requests
type Server struct {
	pb.UnimplementedSimpleBankServer
	config utils.Config
	store db.Store
	tokenMaker token.Maker
	distributor worker.TaskDistributor
}


// NewServer create a new gRPC server and setup routing
func NewServer(config utils.Config, store db.Store, distributor worker.TaskDistributor) (*Server, error) {

	maker, err := token.NewPasetoMaker(config.TokenSymmetryKey)	// or change to JWT
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %v", err)
	}

	server := &Server{
		config:     config,
		store: store,
		tokenMaker: maker,
		distributor: distributor,
	}

	return server, nil
}