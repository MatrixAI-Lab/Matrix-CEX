package matrix

import (
	"MatrixAI-CEX/chain/matrix/matrix_ai"
	"MatrixAI-CEX/common"
	"MatrixAI-CEX/config"
	logs "MatrixAI-CEX/utils/log_utils"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	sendandconfirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"math"
	"strings"
)

var (
	programId                 = solana.MustPublicKeyFromBase58("BvYJnEj64dAT5jrUvgrTJuvPhXRDwJ7SjjPuteycJxAQ")
	mint                      = solana.MustPublicKeyFromBase58("B9pg2pG2vSZWVhe2WEngiCApFCUhwakfnPzpyR3GBKKQ")
	mintDecimals        uint8 = 9
	cexCapital                = solana.MustPrivateKeyFromBase58(config.CEX_CAPITAL_PK)
	cexCapitalPublicKey       = cexCapital.PublicKey()
)

type PlaceOrderParams struct {
	OrderId          string  `binding:"required"`
	Duration         uint32  `binding:"required"`
	Metadata         string  `binding:"required"`
	MachineIdAccount string  `binding:"required"`
	Total            float64 `binding:"required"`
}

type RenewOrderParams struct {
	Duration         uint32  `binding:"required"`
	MachineIdAccount string  `binding:"required"`
	OrderIdAccount   string  `binding:"required"`
	Total            float64 `binding:"required"`
}

func CreateAssociatedAccount(privateKey solana.PrivateKey) (string, error) {
	publicKey := privateKey.PublicKey()

	recent, err := common.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				solana.LAMPORTS_PER_SOL/10,
				cexCapitalPublicKey,
				publicKey,
			).Build(),
			associatedtokenaccount.NewCreateInstruction(
				publicKey,
				publicKey,
				mint,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(cexCapitalPublicKey),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if cexCapitalPublicKey.Equals(key) {
				return &cexCapital
			} else if publicKey.Equals(key) {
				return &privateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== CreateAssociatedAccount Transaction ==================")
	spew.Dump(tx)

	sig, err := sendandconfirm.SendAndConfirmTransaction(
		context.TODO(),
		common.RpcClient,
		common.WsClient,
		tx,
	)
	if err != nil {
		return "", fmt.Errorf("error sending transaction: %v", err)
	}

	logs.Result(fmt.Sprintf("%s completed : %v", "Token.CreateAssociatedAccount", sig.String()))

	return sig.String(), nil
}

func refuel(publicKey solana.PublicKey) []solana.Instruction {
	balance, err := common.RpcClient.GetBalance(context.TODO(), publicKey, rpc.CommitmentFinalized)
	if err != nil {
		return []solana.Instruction{}
	}
	if balance.Value > solana.LAMPORTS_PER_SOL/100 {
		return []solana.Instruction{}
	}

	return []solana.Instruction{
		system.NewTransferInstruction(
			solana.LAMPORTS_PER_SOL/10,
			cexCapitalPublicKey,
			publicKey,
		).Build(),
	}
}

func PlaceOrder(params PlaceOrderParams, privateKey string) (string, error) {
	logs.Normal("Extrinsic : HashrateMarket.PlaceOrder")

	cexCapitalAta, _, err := solana.FindAssociatedTokenAddress(cexCapitalPublicKey, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	buyerPrivateKey := solana.MustPrivateKeyFromBase58(privateKey)
	buyerPublicKey := buyerPrivateKey.PublicKey()
	buyerAta, _, err := solana.FindAssociatedTokenAddress(buyerPublicKey, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	orderIdStr, _ := strings.CutPrefix(params.OrderId, "0x")
	orderIdBytes, err := hex.DecodeString(orderIdStr)
	if err != nil {
		return "", fmt.Errorf("error orderId: %v", err)
	}
	var orderId [16]uint8
	copy(orderId[:], orderIdBytes)

	seedOrder := [][]byte{
		[]byte("order"),
		buyerPublicKey.Bytes(),
		orderIdBytes,
	}
	order, _, err := solana.FindProgramAddress(seedOrder, programId)
	if err != nil {
		return "", fmt.Errorf("error finding program address: %v", err)
	}

	seedVault := [][]byte{
		[]byte("vault"),
		mint.Bytes(),
	}
	vault, _, err := solana.FindProgramAddress(seedVault, programId)
	if err != nil {
		return "", fmt.Errorf("error finding program address: %v", err)
	}

	recent, err := common.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	instructions := refuel(buyerPublicKey)

	instructions = append(instructions,
		token.NewTransferCheckedInstruction(
			uint64(params.Total*math.Pow10(int(mintDecimals))),
			mintDecimals,
			cexCapitalAta,
			mint,
			buyerAta,
			cexCapitalPublicKey,
			[]solana.PublicKey{},
		).Build(),
		matrix_ai.NewPlaceOrderInstruction(
			orderId,
			params.Duration,
			params.Metadata,
			solana.MustPublicKeyFromBase58(params.MachineIdAccount),
			order,
			buyerPublicKey,
			buyerAta,
			vault,
			mint,
			solana.TokenProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
			solana.SystemProgramID,
		).Build(),
	)

	matrix_ai.SetProgramID(programId)
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(cexCapitalPublicKey),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if cexCapitalPublicKey.Equals(key) {
				return &cexCapital
			} else if buyerPublicKey.Equals(key) {
				return &buyerPrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== PlaceOrder Transaction ==================")
	spew.Dump(tx)

	sig, err := sendandconfirm.SendAndConfirmTransaction(
		context.TODO(),
		common.RpcClient,
		common.WsClient,
		tx,
	)
	if err != nil {
		spew.Dump(err)
		return "", fmt.Errorf("error sending transaction: %v", err)
	}

	logs.Result(fmt.Sprintf("%s completed : %v", "HashrateMarket.PlaceOrder", sig.String()))

	return sig.String(), nil
}

func RenewOrder(params RenewOrderParams, privateKey string) (string, error) {
	logs.Normal("Extrinsic : HashrateMarket.RenewOrder")

	cexCapitalAta, _, err := solana.FindAssociatedTokenAddress(cexCapitalPublicKey, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	buyerPrivateKey := solana.MustPrivateKeyFromBase58(privateKey)
	buyerPublicKey := buyerPrivateKey.PublicKey()
	buyerAta, _, err := solana.FindAssociatedTokenAddress(buyerPublicKey, mint)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	seedVault := [][]byte{
		[]byte("vault"),
		mint.Bytes(),
	}
	vault, _, err := solana.FindProgramAddress(seedVault, programId)
	if err != nil {
		return "", fmt.Errorf("error finding program address: %v", err)
	}

	recent, err := common.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	instructions := refuel(buyerPublicKey)

	instructions = append(instructions,
		token.NewTransferCheckedInstruction(
			uint64(params.Total*math.Pow10(int(mintDecimals))),
			mintDecimals,
			cexCapitalAta,
			mint,
			buyerAta,
			cexCapitalPublicKey,
			[]solana.PublicKey{},
		).Build(),
		matrix_ai.NewRenewOrderInstruction(
			params.Duration,
			solana.MustPublicKeyFromBase58(params.MachineIdAccount),
			solana.MustPublicKeyFromBase58(params.OrderIdAccount),
			buyerPublicKey,
			buyerAta,
			vault,
			mint,
			solana.TokenProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
		).Build(),
	)

	matrix_ai.SetProgramID(programId)
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(cexCapitalPublicKey),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if cexCapitalPublicKey.Equals(key) {
				return &cexCapital
			} else if buyerPublicKey.Equals(key) {
				return &buyerPrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== RenewOrder Transaction ==================")
	spew.Dump(tx)

	sig, err := sendandconfirm.SendAndConfirmTransaction(
		context.TODO(),
		common.RpcClient,
		common.WsClient,
		tx,
	)
	if err != nil {
		spew.Dump(err)
		return "", fmt.Errorf("error sending transaction: %v", err)
	}

	logs.Result(fmt.Sprintf("%s completed : %v", "HashrateMarket.PlaceOrder", sig.String()))

	return sig.String(), nil
}
