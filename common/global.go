package common

import (
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Db        *gorm.DB
	Rdb       *redis.Client
	RpcClient *rpc.Client
	WsClient  *ws.Client
)
