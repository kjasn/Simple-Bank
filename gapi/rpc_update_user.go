package gapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	supportedAuthorizationType = "bearer"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// check payload first
	authPayload, err := server.authorization(ctx, []db.UserRole{db.UserRoleDepositor, db.UserRoleBanker})
	if err != nil {
		return nil, unAuthenticatedError(err)
	}
	
	// only employees can modify others' profile
	if authPayload.Username != string(db.UserRoleBanker) && authPayload.Username != req.GetUsername() {
		return nil, fmt.Errorf("can not update other's info")
	}

	// validate request parameters
	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	} 


	arg := db.UpdateUserParams {
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid: req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid: req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "fail to hash password: %s", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid: true,
		}
	}


	user, err := server.store.UpdateUser(ctx, arg)	// do not update username
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "fail to update user: %s", err)
	}

	ret := &pb.UpdateUserResponse{
		User: convertUser(&user),
	}

	return ret, nil
}


func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violation []*errdetails.BadRequest_FieldViolation){
	// username is required
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}

	// password, full name and email are optional 
	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violation = append(violation, fieldViolation("password", err))
		}
	}

	if req.FullName != nil {
		if err := val.ValidateFullName(req.GetFullName()); err != nil {
			violation = append(violation, fieldViolation("full_name", err))
		}
	}

	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violation = append(violation, fieldViolation("email", err))
		}
	}

	return violation
}
