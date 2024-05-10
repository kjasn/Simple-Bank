package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/token"
	"github.com/kjasn/simple-bank/utils"
)

// Server servers HTTP requests
type Server struct {
	config utils.Config
	store db.Store
	tokenMaker token.Maker
	router *gin.Engine
}


// NewServer create a new HTTP server and setup routing
func NewServer(config utils.Config, store db.Store) (*Server, error) {

	maker, err := token.NewPasetoMaker(config.TokenSymmetryKey)	// or change to JWT
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %v", err)
	}

	server := &Server{
		config:     config,
		store: store,
		tokenMaker: maker,
	}


	// set up router
	server.setUpRouter()
	
	// register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)	// name and validator func
	}

	return server, nil
} 

// setUpRouter 
func (server *Server) setUpRouter() {
	router := gin.Default()
	// add router
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/token/access_token", server.renewAccessToken)
	
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.GET("/users/:username", server.getUser)
	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}


// Start  run the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error" : err.Error()}
}