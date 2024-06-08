package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/kjasn/simple-bank/db/sqlc"
)

// add 2 queue type
const (
	QueueMain = "main"
	QueueDefault = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(
		ctx context.Context,
		task *asynq.Task,
	) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store db.Store	
}


func NewRedisTaskProcessor (redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int {
				QueueMain: 10, 
				QueueDefault: 5,
			},
		},	
	)

	return &RedisTaskProcessor {
		server: server,
		store: store,
	}
}


func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// register processor
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}