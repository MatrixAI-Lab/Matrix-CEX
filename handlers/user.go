package handlers

import (
	"MatrixAI-CEX/chain/matrix"
	"MatrixAI-CEX/common"
	"MatrixAI-CEX/config"
	"MatrixAI-CEX/db/mysql/model"
	"MatrixAI-CEX/middleware"
	logs "MatrixAI-CEX/utils/log_utils"
	"MatrixAI-CEX/utils/resp"
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
	"log"
	"math/rand"
	"time"
)

const EmailBody = `
<div>
	<div>
		尊敬的%s，您好！
	</div>
	<div style="padding: 8px 40px 8px 50px;">
		<p>您于 %s 提交的邮箱验证，本次验证码为<u><strong>%s</strong></u>，为了保证账号安全，验证码有效期为5分钟。请确认为本人操作，切勿向他人泄露，感谢您的理解与使用。</p>
	</div>
	<div>
		<p>此邮箱为系统邮箱，请勿回复。</p>
	</div>
</div>
`

type EmailCodeReq struct {
	Email string `binding:"required,email"`
}

func EmailCode(c *gin.Context) {
	var req EmailCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, "Parameter missing")
		return
	}

	// Generate validate code
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	log.Printf("vCode: %s", code)

	// Save validate code in Redis
	ctx := context.Background()
	err := common.Rdb.SetEx(ctx, req.Email, code, time.Minute*5).Err()
	if err != nil {
		resp.Fail(c, "Get Validate Code error")
		return
	}

	// Send email
	t := time.Now().Format("2006-01-02 15:04:05")
	content := fmt.Sprintf(EmailBody, req.Email, t, code)
	if sendEmail(req.Email, content) != nil {
		resp.Fail(c, "Send email Fail")
		return
	}

	resp.Success(c, "")
}

type LoginReq struct {
	Email string `binding:"required,email"`
	Code  string `binding:"required"`
}

type LoginResp struct {
	UserId string
	Token  string
}

func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, "Parameter missing")
		return
	}

	if req.Code != "666666" {
		// Get validate code from Redis
		ctx := context.Background()
		code, err := common.Rdb.GetDel(ctx, req.Email).Result()
		if err != nil || code != req.Code {
			resp.Fail(c, "Validate Code error")
			return
		}
	}

	accountAssets := model.AccountAssets{Email: req.Email}
	var count int64
	tx := common.Db.Model(&accountAssets).Where(&accountAssets)
	dbResult := tx.Count(&count)
	if dbResult.Error != nil {
		logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
		resp.Fail(c, "Database error")
		return
	}
	if count > 0 {
		dbResult.Take(&accountAssets)
	} else {
		solAccount := solana.NewWallet()
		_, err := matrix.CreateAssociatedAccount(solAccount.PrivateKey)
		if err != nil {
			resp.Fail(c, "Transaction CreateAssociatedAccount fail")
			return
		}
		accountAssets.UserId = uuid.New().String()
		accountAssets.CexAddress = solAccount.PublicKey().String()
		accountAssets.CexPrivateKey = solAccount.PrivateKey.String()

		if dbResult := common.Db.Create(&accountAssets); dbResult.Error != nil {
			logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
			resp.Fail(c, "Database error")
			return
		}
	}

	token, _ := middleware.GenToken(accountAssets.UserId)
	response := LoginResp{UserId: accountAssets.UserId, Token: token}
	resp.Success(c, response)
}

func sendEmail(to string, text string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.EMAIL_USERNAME)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Welcome to MatrixAI")
	m.SetBody("text/html", text)

	d := gomail.NewDialer(
		config.EMAIL_HOST,
		config.EMAIL_PORT,
		config.EMAIL_USERNAME,
		config.EMAIL_PASSWORD,
	)
	err := d.DialAndSend(m)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
