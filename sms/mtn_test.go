package sms

import (
	"testing"
)

func TestMtn(t *testing.T) {

	tel := "135xxxxyyyy"

	SmsSender := &MtnSender{
		Host:  "shmtn.b2m.cn",
		AppID: "",
		Pwd:   "",
		Tpls: []MtnTpl{
			{
				Type:   1,
				Params: "【快豹】您的验证码为 %s",
				Len:    1,
			},
		},
	}

	err := SmsSender.Send(1, tel, "123456")
	if err != nil {
		t.Error(err)
	}

}
