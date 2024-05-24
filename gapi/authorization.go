package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/kjasn/simple-bank/token"
	"google.golang.org/grpc/metadata"
)


func (server *Server) authorization (ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx) 
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	value := md.Get(authorizationHeader)
	if len(value) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := value[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != supportedAuthorizationType {
		return nil, fmt.Errorf("unsupported token type: %s", authType)
	}

	token := fields[1]
	payload, err := server.tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}