package gapi

import (
	"context"
	"database/sql"

	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)


func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	// first validate request parameters
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	} 

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "fail to found user: %s", err)
	}

	err = utils.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "wrong password: %s", err)
	}
	
	// create access token 
	accessToken, accessTokenPayload, err := server.tokenMaker.CreateToken(
		req.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to create access token: %s", err)
	}

	// create refresh token after access token
	refreshToken, refreshTokenPayload, err := server.tokenMaker.CreateToken(
		req.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to create refresh token: %s", err)
	}


	// store the refresh token into session table
	mtdt := server.extractMetadata(ctx)
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID: refreshTokenPayload.ID,
		Username: user.Username,
		RefreshToken : refreshToken,
		UserAgent : mtdt.UserAgent,
		ClientIp : mtdt.ClientIP,
		IsBlocked : false, 
		ExpiresAt: refreshTokenPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to create session: %s", err)
	}


	ret := &pb.LoginUserResponse {
		SessionId: session.ID.String(),
		AccessToken: accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessTokenPayload.ExpiredAt),
		RefreshToken: refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshTokenPayload.ExpiredAt),
		User: convertUser(&user),
	}
	return ret, nil
}


func validateLoginUserRequest(req *pb.LoginUserRequest) (violation []*errdetails.BadRequest_FieldViolation){
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violation = append(violation, fieldViolation("password", err))
	}

	return violation
}
