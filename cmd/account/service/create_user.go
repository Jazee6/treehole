package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Jazee6/treehole/cmd/account/dao"
	"github.com/Jazee6/treehole/cmd/account/model"
	"github.com/Jazee6/treehole/cmd/account/rpc"
	"github.com/Jazee6/treehole/pkg/rpcs"
	"github.com/Jazee6/treehole/pkg/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type AccountService struct{}

func (c *AccountService) SendCaptcha(ctx context.Context, request *rpc.SendCaptchaRequest) (*rpc.SendCaptchaResponse, error) {
	err := Recaptcha(request.Token)
	if err != nil {
		if err.Error() == "recaptcha failed" {
			return &rpc.SendCaptchaResponse{
				Code: rpcs.Code_ErrRecaptchaErr,
			}, nil
		}
		return nil, err
	}

	// 验证邮箱是否已经注册
	q := dao.Q.User
	user, err := q.Where(q.Email.Eq(request.Email)).Take()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if user != nil {
		return &rpc.SendCaptchaResponse{
			Code: rpcs.Code_ErrUserExist,
		}, nil
	}

	// 验证码限流
	result, err := r.Get(ctx, "email").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if err == redis.Nil {
		now := time.Now()
		year, month, day := now.Date()
		err := r.Set(ctx, "email", 0, time.Date(year, month, day, 24, 0, 0, 0, time.Local).Sub(time.Now())).Err()
		if err != nil {
			return nil, err
		}
		result = "0"
	}
	re, err := strconv.Atoi(result)
	if err != nil {
		return nil, err
	}
	if emailMax < re {
		return &rpc.SendCaptchaResponse{
			Code: rpcs.Code_ErrEmailLimit,
		}, nil
	}
	err = r.Incr(ctx, "email").Err()
	if err != nil {
		return nil, err
	}
	code := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(999999))
	err = r.Set(ctx, request.Email, code, time.Minute*time.Duration(captchaExpire)).Err()
	if err != nil {
		return nil, err
	}
	err = utils.SendMail(request.Email, "【"+name+"】"+"验证码", os.Expand(captchaContent, func(s string) string {
		switch s {
		case "code":
			return code
		}
		return "{" + s + "}"
	}))
	if err != nil {
		return nil, err
	}

	return &rpc.SendCaptchaResponse{
		Code: rpcs.Code_Success,
	}, nil
}

func (c *AccountService) AccountRegister(ctx context.Context, request *rpc.RegisterRequest) (*rpc.RegisterResponse, error) {
	result, err := r.Get(ctx, request.Email).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if err == redis.Nil {
		return &rpc.RegisterResponse{
			Code: rpcs.Code_ErrCaptchaNil,
		}, nil
	}
	if result != request.Captcha {
		return &rpc.RegisterResponse{
			Code: rpcs.Code_ErrCaptchaErr,
		}, nil
	}

	q := dao.Q.User
	user, err := q.Where(q.Email.Eq(request.Email)).Take()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if user != nil {
		return &rpc.RegisterResponse{
			Code: rpcs.Code_ErrUserExist,
		}, nil
	}
	s := sha256.New()
	s.Write([]byte(request.Password + salt))
	usr := &model.User{
		NickName:  request.Nickname,
		Email:     request.Email,
		Password:  hex.EncodeToString(s.Sum(nil)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = q.Create(usr)
	if err != nil {
		return nil, err
	}
	tk, err := utils.GenToken(*usr)
	if err != nil {
		return nil, err
	}
	return &rpc.RegisterResponse{
		Code:  rpcs.Code_Success,
		Token: tk,
	}, nil
}
