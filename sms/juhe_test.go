package sms

import (
	"testing"
)

func TestJuheSend(t *testing.T) {
	tplid := "test"
	key := "test"
	tel := "135xxxxyyyy"

	juheSender := JuheSender{
		Key: key,
		Tpls: []JuheTpl{
			{
				Type:   0,
				ID:     tplid,
				Params: "#code#=%s",
				Len:    1,
			},
		},
	}

	err := juheSender.Send(0, tel, "123456")
	if err != nil {
		t.Error(err)
	}
}
