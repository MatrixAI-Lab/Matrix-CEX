package config

import "github.com/gagliardetto/solana-go/rpc"

const RPC = rpc.DevNet_RPC  // 开发链
const WsRPC = rpc.DevNet_WS // 开发链

const CEX_CAPITAL_POOL = "CWub1ia9Bispahk8J3FExFDaPzAk1jYXmAzTiTdxSVLp"

const URL = IP + ":" + POST
const IP = "172.18.232.45"
const POST = "8080"

// const DSN = "root:root1234@tcp(localhost:3306)/exchange_db?charset=utf8mb4&parseTime=True&loc=Local"
const DSN = DB_USER + ":" + DB_PASSWORD + "@tcp(localhost:3306)/" + DB_NAME + "?charset=utf8mb4&parseTime=True&loc=Local"
const DB_NAME = "exchange_db"
const DB_USER = "root"
const DB_PASSWORD = "root1234"