package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"
const LenOfSecretCode = 32

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}


func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	// marshal task into slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	// create task
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	// enqueue task
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// log out
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}


func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(
	ctx context.Context,
	task *asynq.Task,
) error {
	// extra payload
	var payload PayloadSendVerifyEmail
	
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
    }

	// retrieve user record from db
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		// user not exists
		// if err == sql.ErrNoRows {
        	// return fmt.Errorf("user not exists: %w", asynq.SkipRetry)
		// }
        return fmt.Errorf("failed to get user from database: %w", err)
	}


	verifyEmail, err := processor.store.VerifyEmail(ctx, db.VerifyEmailParams{
		SecretCode: utils.RandomString(LenOfSecretCode),
		Username: user.Username,
		Email: user.Email,
	})
	if err != nil {
        return fmt.Errorf("failed to create verify email: %w", err)
	}

	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?verify_email_id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)

	content := fmt.Sprintf(`<h1> Hello %s, welcome to simple bank ! </h1>
	Please <a href="%s"> click here </a> to verify your email address. </br>
	If you did not register a new account, please ignore it! </br>
	`, user.Username, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SendMail(subject, content, to, nil, nil, nil)
	if err != nil {
        return fmt.Errorf("failed to send verify email: %w", err)
	}

	// log out
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	 
	return nil
}