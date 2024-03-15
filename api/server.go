package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/kjasn/simple-bank/db/sqlc"
)

// Server servers HTTP requests
type Server struct {
	store *db.Store
	router *gin.Engine
}


// NewServer create a new HTTP server and setup routing
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add router
	router.POST("/accounts", server.createAccount)

	server.router = router
	return server
} 

// Start  run the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error" : err.Error()}
}