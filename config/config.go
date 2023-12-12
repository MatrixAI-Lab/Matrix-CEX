package config

import "github.com/gagliardetto/solana-go/rpc"

const RPC = rpc.DevNet_RPC  // 开发链
const WsRPC = rpc.DevNet_WS // 开发链

const CEX_CAPITAL_POOL = "CWub1ia9Bispahk8J3FExFDaPzAk1jYXmAzTiTdxSVLp"
const CEX_CAPITAL_PK = "2py8uvpGazsSm9zmFSjkx6q37C5fmvTkg2k4hy9H2aaWYiXqs9WTK1aq15W1Y3Dj7vF8KCYMpMQ11dTfQViaxnUC"

const URL = DEBUG_IP + ":" + DEBUG_POST

// 本机 develop
const DEV_IP = "172.18.232.45"
const DEV_POST = "8080"
// 3.27.91.240 debug
const DEBUG_IP = "0.0.0.0"
const DEBUG_POST = "8099"

// const DSN = "root:root1234@tcp(localhost:3306)/exchange_db?charset=utf8mb4&parseTime=True&loc=Local"
const DSN = DEBUG_DB_USER + ":" + DEBUG_DB_PASSWORD + "@tcp(localhost:3306)/" + DB_NAME + "?charset=utf8mb4&parseTime=True&loc=Local"
const DB_NAME = "exchange_db"

// 本机 develop
const DEV_DB_USER = "root"
const DEV_DB_PASSWORD = "root1234"
// 3.27.91.240 debug
const DEBUG_DB_USER = "root"
const DEBUG_DB_PASSWORD = "Zxcvbn2023@"

const DEBUG_PASSWORD = "MatrixAI"
