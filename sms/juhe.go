package sms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// JuheSender 聚合接口
//
// r.Post("/sms", ws.Sms{}.Handler(ws.JuheSender{
// 	TplID: "",
// 	Key:   "",
// }))
//
type JuheSender struct {
	Key  string
	Tpls []JuheTpl
}

// JuheTpl 短信模板
type JuheTpl struct {
	Type   uint   // 类型
	ID     string // 短信ID
	Params string // 参数模板
	Len    int    // 参数长度
}

// ValidLen 验证长度
func (s JuheTpl) ValidLen(n int) error {
	if n != s.Len {
		return errors.New("短信参数错误")
	}
	return nil
}

// Parse 解析处理
func (s JuheTpl) Parse(data ...interface{}) string {
	// 解析模板 模板 验证对应参数是否对照
	return fmt.Sprintf(s.Params, data...)
}

type juheReturn struct {
	Code   int    `json:"error_code"`
	Reason string `json:"reason"`
}

// Tpl 模板
func (s JuheSender) Tpl(smsType uint) (SmsTpl, error) {
	for _, v := range s.Tpls {
		if v.Type == smsType {
			return v, nil
		}
	}

	return nil, errors.New("短信模板不存在，请更换短信类型！")
}

// Send 发送消息
func (s JuheSender) Send(smsType uint, tel string, data ...interface{}) error {

	tpli, err := s.Tpl(smsType)
	if err != nil {
		return err
	}
	tpl, _ := tpli.(JuheTpl)

	v := url.Values{}
	v.Add("mobile", tel)
	v.Add("key", s.Key)
	v.Add("tpl_id", tpl.ID)
	// v.Add("tpl_value", fmt.Sprintf("#code#=%s", data...))
	v.Add("tpl_value", tpl.Parse(data...))

	urlpost := fmt.Sprintf("https://v.juhe.cn/sms/send?%s", v.Encode())

	req, _ := http.NewRequest("POST", urlpost, nil)
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New("短信服务商连接错误！")
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return errors.New("短信服务商连接错误！")
	}
	re := &juheReturn{}
	if err := json.Unmarshal(body, re); err == nil {
		if re.Code == 0 {
			return nil
		} else {
			return errors.New(re.Reason)
		}
	}
	// fmt.Println(res)
	// fmt.Println(string(body))

	return errors.New("短信发送失败")
}
