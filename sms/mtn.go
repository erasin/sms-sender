package sms

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MtnSender 亿美软通
// 文档： http://mtn.b2m.cn:8080/static/doc/onlysms.html
type MtnSender struct {
	Host  string
	AppID string // appid
	Pwd   string
	Tpls  []MtnTpl
}

// MtnTpl 短信模板
type MtnTpl struct {
	Type   uint   // 类型
	ID     string // 短信ID
	Params string // 参数模板
	Len    int    // 参数长度
}

// ValidLen 验证长度
func (s MtnTpl) ValidLen(n int) error {
	if n != s.Len {
		return errors.New("短信参数错误")
	}
	return nil
}

// Parse 解析处理
func (s MtnTpl) Parse(data ...interface{}) string {
	// 解析模板 模板 验证对应参数是否对照
	return fmt.Sprintf(s.Params, data...)
}

// Tpl 模板
func (s MtnSender) Tpl(smsType uint) (SmsTpl, error) {
	for _, v := range s.Tpls {
		if v.Type == smsType {
			return v, nil
		}
	}

	return nil, errors.New("短信模板不存在，请更换短信类型！")
}

type mtnReturn struct {
	Code        string `json:"Code"` // 状态码
	Mobile      string `json:"mobile"`
	SmsID       string `json:"smsId"`
	CustomSmsID string `json:"customSmsId"`
}

// Send 发送消息
func (s MtnSender) Send(smsType uint, tel string, data ...interface{}) error {
	tpli, err := s.Tpl(smsType)
	if err != nil {
		return err
	}
	tpl, _ := tpli.(MtnTpl)

	code := tpl.Parse(data...)

	t := time.Now().In(BeijingLocation).UTC().Format("20060102150405")
	sign := md5Crypt(s.AppID, s.Pwd, t)

	var params = map[string]string{}
	params["appId"] = s.AppID
	params["timestamp"] = t
	params["mobiles"] = tel
	// params["sign"] = sign
	// params["content"] = url.QueryEscape(code)
	params["content"] = specialUrlEncode(code)
	params["timerTime"] = ""
	params["customSmsId"] = ""
	params["extendedCode"] = ""

	var paramsStr string
	for k, v := range params {
		paramsStr += "&" + specialUrlEncode(k) + "=" + v
	}

	urlStr := "http://" + s.Host + "/simpleinter/sendSMS?sign=" + specialUrlEncode(sign) + paramsStr

	fmt.Printf("%s\n", urlStr)

	req, _ := http.NewRequest("GET", urlStr, nil)
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
	re := &mtnReturn{}
	if err := json.Unmarshal(body, re); err == nil {
		if re.Code != "SUCCESS" {
			return errors.New("无法发送短信")
		}
	}

	return nil
}

// md5Crypt 加密
func md5Crypt(str string, salt ...interface{}) (CryptStr string) {
	if l := len(salt); l > 0 {
		slice := make([]string, l+1)
		str = fmt.Sprintf(str+strings.Join(slice, "%v"), salt...)
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
