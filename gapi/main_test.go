package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/token"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
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


func buildContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, role db.UserRole, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", supportedAuthorizationType, accessToken)
	md := metadata.MD {
		authorizationHeader: []string{bearerToken},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}