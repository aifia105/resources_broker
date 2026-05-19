package queue

import "github.com/hibiken/asynq"

func NewAsynqClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})
}
