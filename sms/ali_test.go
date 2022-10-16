package sms

import (
	"testing"
)

func TestAliSend(t *testing.T) {

	secret := "test"
	key := "test"
	tplid := "SMS_xxx"
	tel := "135xxxxyyyy"
	// 正式驱动
	aliSender := AliSender{
		Key:      key,
		Secret:   secret,
		SignName: "xx科技",
		Tpls: []AliTpl{
			{
				Type:   0,
				ID:     tplid,
				Params: `{"code":"%s"}`,
				Len:    1,
			},
		},
	}

	err := aliSender.Send(0, tel, "123456")
	if err != nil {
		t.Error(err)
	}

}
