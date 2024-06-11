package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/kjasn/simple-bank/mail"
	"github.com/kjasn/simple-bank/utils"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

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

	config, err := utils.LoadConfig("../")
	if err != nil {
        return fmt.Errorf("failed to load configuration file: %w", err)
	}

	netEaseMail := mail.NewNetEaseMailSender(
		config.EmailSenderName, 
		config.EmailSenderAddress, 
		config.EmailSenderPassword,
	)		

	subject := "A Test Email"
	content := `
	<h1> Hello World </h1>
	<p> This is a test mail </p>
	`
	to := []string{"mail-address@example.com"}
	err = netEaseMail.SendMail(subject, content, to, nil, nil, nil)
	if err != nil {
        return fmt.Errorf("failed to send email: %w", err)
	}

	// log out
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	 
	return nil
}