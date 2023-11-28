package conn

import (
	"MatrixAI-CEX/config"
	"MatrixAI-CEX/db/mysql/model"
	logs "MatrixAI-CEX/utils/log_utils"
	"context"
	"fmt"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	sendandconfirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type Conn struct {
	RpcClient *rpc.Client
	WsClient  *ws.Client
}

func NewConn() (*Conn, error) {

	rpcClient := rpc.New(config.RPC)
	wsClient, err := ws.Connect(context.Background(), config.WsRPC)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		RpcClient: rpcClient,
		WsClient:  wsClient,
	}

	return conn, nil
}

func (conn *Conn) GetBalance(publicKey string) (string, error) {
	pubKey := solana.MustPublicKeyFromBase58(publicKey)
	out, err := conn.RpcClient.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return "", err
	}
	var lamportsOnAccount = new(big.Float).SetUint64(uint64(out.Value))
	// Convert lamports to sol:
	var solBalance = new(big.Float).Quo(lamportsOnAccount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	return solBalance.Text('f', 10), nil
}

func (conn *Conn) RechargeSol(accountAssets model.AccountAssets) (string, uint64, error) {

	pubKey := solana.MustPublicKeyFromBase58(accountAssets.CexAddress)
	out, err := conn.RpcClient.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return "", 0, err
	}

	amount := uint64(out.Value)

	logs.Normal(fmt.Sprintf("CEX SOL balance: %d", amount))

	if amount >= solana.LAMPORTS_PER_SOL {
		accountFrom, err := solana.PrivateKeyFromBase58(accountAssets.CexPrivateKey)
		if err != nil {
			return "", 0, err
		}
		accountTo := solana.MustPublicKeyFromBase58(config.CEX_CAPITAL_POOL)

		recent, err := conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
		if err != nil {
			return "", 0, err
		}

		amount = amount - 1000000

		tx, err := solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					amount,
					accountFrom.PublicKey(),
					accountTo,
				).Build(),
			},
			recent.Value.Blockhash,
			solana.TransactionPayer(accountFrom.PublicKey()),
		)
		if err != nil {
			return "", 0, err
		}

		_, err = tx.Sign(
			func(key solana.PublicKey) *solana.PrivateKey {
				if accountFrom.PublicKey().Equals(key) {
					return &accountFrom
				}
				return nil
			},
		)
		if err != nil {
			return "", 0, fmt.Errorf("unable to sign transaction: %w", err)
		}

		// Send transaction, and wait for confirmation:
		sig, err := sendandconfirm.SendAndConfirmTransaction(
			context.TODO(),
			conn.RpcClient,
			conn.WsClient,
			tx,
		)
		if err != nil {
			return "", 0, fmt.Errorf("unable to send transaction: %w", err)
		}
		return sig.String(), amount, nil
	}
	return "", 0, nil
}

func (conn *Conn) Withdraw(toAddress string, toAmount uint64) (string, error) {

	pubKey := solana.MustPublicKeyFromBase58(config.CEX_CAPITAL_POOL)
	out, err := conn.RpcClient.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return "", err
	}

	amount := uint64(out.Value)

	if amount >= toAmount {
		accountFrom, err := solana.PrivateKeyFromBase58(config.CEX_CAPITAL_PK)
		if err != nil {
			return "", err
		}
		accountTo := solana.MustPublicKeyFromBase58(toAddress)

		recent, err := conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
		if err != nil {
			return "", err
		}

		tx, err := solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					toAmount,
					accountFrom.PublicKey(),
					accountTo,
				).Build(),
			},
			recent.Value.Blockhash,
			solana.TransactionPayer(accountFrom.PublicKey()),
		)
		if err != nil {
			return "", err
		}

		_, err = tx.Sign(
			func(key solana.PublicKey) *solana.PrivateKey {
				if accountFrom.PublicKey().Equals(key) {
					return &accountFrom
				}
				return nil
			},
		)
		if err != nil {
			return "", fmt.Errorf("unable to sign transaction: %w", err)
		}

		// Send transaction, and wait for confirmation:
		sig, err := sendandconfirm.SendAndConfirmTransaction(
			context.TODO(),
			conn.RpcClient,
			conn.WsClient,
			tx,
		)
		if err != nil {
			return "", fmt.Errorf("unable to send transaction: %w", err)
		}
		return sig.String(), nil
	}
	return "", fmt.Errorf("CEX SOL insufficient balance")
}
