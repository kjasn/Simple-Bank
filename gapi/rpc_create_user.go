package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/val"
	"github.com/kjasn/simple-bank/worker"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// first validate request parameters
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	} 

	hashedPassword, err := utils.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to hash password: %s", err)
	}

	arg := db.CreateUserTxParams {
		CreateUserParams: db.CreateUserParams{
			Username: req.GetUsername(),
			Role: db.UserRole(req.GetRole()),
			HashedPassword: hashedPassword,
			FullName: req.GetFullName(),
			Email: req.GetEmail(),
		},
		AfterCreated: func(user db.User) error{
			payload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}

			opts := []asynq.Option {
				asynq.MaxRetry(10),
				asynq.ProcessIn(5 * time.Second),
				asynq.Queue(worker.QueueMain),
			}
			return server.distributor.DistributeTaskSendVerifyEmail(ctx, payload, opts...)
		},
	}

	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}

		return nil, status.Errorf(codes.Internal, "fail to create user: %s", err)
	}


	ret := &pb.CreateUserResponse{
		User: convertUser(&txResult.User),
	}

	return ret, nil
}


func validateCreateUserRequest(req *pb.CreateUserRequest) (violation []*errdetails.BadRequest_FieldViolation){
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}

	if err := val.ValidUserRole(req.GetRole()); err != nil {
		violation = append(violation, fieldViolation("role", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violation = append(violation, fieldViolation("password", err))
	}

	if err := val.ValidateFullName(req.GetFullName()); err != nil {
		violation = append(violation, fieldViolation("full_name", err))
	}

	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violation = append(violation, fieldViolation("email", err))
	}

	return violation
}
