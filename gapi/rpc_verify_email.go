package gapi

import (
	"context"

	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)



func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	// validate request parameters first
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		VerifyEmailId: req.VerifyEmailId,	
		SecretCode: req.SecretCode,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to verify email")
	}

	rsp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}

	return rsp, nil
}


func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violation []*errdetails.BadRequest_FieldViolation){
	if err := val.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violation = append(violation, fieldViolation("secret_code", err))
	}

	if err := val.ValidateVerifyEmailId(req.GetVerifyEmailId()); err != nil {
		violation = append(violation, fieldViolation("verify_email_id", err))
	}

	return violation
}
