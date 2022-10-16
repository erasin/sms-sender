# 短信发送器

sms-sender/sms 是使用 golang 实现短信发送器。

目前实现下面三个服务商。

- [阿里短信](https://help.aliyun.com/product/44282.html)
- [聚合短信](https://www.juhe.cn/docs/api/id/54)
- [亿美软通](http://mtn.b2m.cn:8080/static/doc/onlysms.html)


## 使用方式

### 1. 定义驱动器

```go
package sender

import "github.com/erasin/sms-sender/sms"

// SmsSender 短信发送着额
var SmsSender sms.SmsSender

// 定义短信类型 enum
const (
	SmsTypeLogin uint = iota + 1 // 短信登录验证码
	SmsTypeReg                   // 注册用
	SmsTypeRepass                // 找回密码
)

func initSms() {
	// 测试用驱动
	if Conf.IsDebugging() {
		SmsSender = &sms.SmsTestSender{}
	} else {
		// 正式驱动
		SmsSender = &sms.AliSender{
			Key:      Conf.Sms.Key,
			Secret:   Conf.Sms.Secret,
			SignName: Conf.Sms.Sign,
			Tpls: []sms.AliTpl{
				{
					Type:   SmsTypeLogin,
					ID:     "SMS_1111111111",
					Params: `{"code":"%s"}`,
					Len:    1,
				},
				{
					Type:   SmsTypeRepass,
					ID:     "SMS_2222222222",
					Params: `{"code":"%s"}`,
					Len:    1,
				},
			},
		}
		// sms.JuheSender
		// sms.MtnSender
	}
}
```

### 2. 使用短信发送器

```go
// ...
	if err := C.SmsSender.Send(C.SmsTypeLogin, tel, code); err != nil {
		c.JSON(http.StatusOK, err.Error())
		return
	}
// ...
```

### 3. 中间件

将发送的短信记录到redis中,方便后续的操作。

> 下面的例子具体会另开一个整合项目进行展示

将中间件实例化

```go
package action

import (
	"errors"
	"net/http"
	"github.com/erasin/sms-sender/middleware"

    // 伪代码
	"project/core"
	"project/model"
)

// SmsCode 短信验证中间件
var SmsCode *mw.SmsCode

func initSms() {

	SmsCode = middleware.NewSmsCode(C.Log, C.RedisPoolLink, SmsChecker{})

	if !C.Conf.IsRelease() {
		SmsCode.Debug = true
	}
}

// SmsChecker 短信验证器
type SmsChecker struct{}

// Get 获取当前登陆账号的手机号
func (s SmsChecker) Get(smsType uint, r *http.Request) (tel string, err error) {
    // do some
	return "", nil
}

// HaveType 验证短信验证码类型
func (s SmsChecker) HaveType(smsType uint) error {
	// 短信验证码类型验证
	switch smsType {
	case C.SmsTypeLogin, C.SmsTypeRepass:
		return nil
	}
	return errors.New("非法的短信请求类型")
}

// Valid 验证账号有效性
// 默认情况发送 nil
func (s SmsChecker) Valid(smsType uint, tel string) error {
	// 验证身份
	account := &M.Account{}
	account.Tel = tel
	account.Get()

	switch smsType {
	case C.SmsTypeLogin, C.SmsTypeRepass:
		if account.ID == 0 {
			return errors.New("没有找到有效的用户")
		}
		if !account.Status {
			return errors.New("账户已被禁用，请联系管理员处理")
		}
	}

	return nil
}

// 这里一般为加盐处理
func (s SmsChecker) RdbKey(key ...string) string {
	return C.RdbKey(key...)
}
```

使用

```go

// 短信处理
r.POST("/sms", SmsCode.Handle(C.SmsSender))
r.POST("/sms/valid", SmsCode.HandleValid)

// gin 路由
r.POST("/login/tel", SmsCode.SmsValidCtx(C.SmsTypeLogin), a.loginByTel)
r.POST("/repass", SmsCode.SmsValidCtx(C.SmsTypeRepass), a.repass)
     
```