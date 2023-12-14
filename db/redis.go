package db

import (
	"MatrixAI-CEX/common"
	"github.com/redis/go-redis/v9"
)

func InitRedis() {
	common.Rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
}
