package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)


func newTestServer(t *testing.T, store db.Store) *Server {
	config := utils.Config {
		TokenSymmetryKey: utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	// set test mode to make the logs clearly
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}