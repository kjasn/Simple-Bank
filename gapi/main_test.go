package gapi

import (
	"testing"
	"time"

	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/worker"
	"github.com/stretchr/testify/require"
)


func newTestServer(t *testing.T, store db.Store, distributor worker.TaskDistributor) *Server {
	config := utils.Config {
		TokenSymmetryKey: utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, distributor)
	require.NoError(t, err)
	return server
}