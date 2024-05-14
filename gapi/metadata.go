package gapi

import (
	"context"
	"log"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader = "user-agent"
	xForwardedForHeader = "x-forwarded-for"
)

type Metadata struct {
	UserAgent 	string
	ClientIP 	string
}

func (Server *Server) extractMetadata(ctx context.Context) *Metadata{
	mtdt := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("metadata: %+v\n", md)

		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtdt.ClientIP= clientIPs[0]
		}
	}

	// extract client ip from gRPC request from context by peer
	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = p.Addr.String()
	}
	return mtdt
}