package sms

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

// AliSender 阿里短信推送
type AliSender struct {
	Key      string
	Secret   string
	SignName string
	Tpls     []AliTpl
}

// AliTpl 短信模板
type AliTpl struct {
	Type   uint   // 类型
	ID     string // 短信ID
	Params string // 参数模板
	Len    int    // 参数长度
}

// ValidLen 验证长度
func (s AliTpl) ValidLen(n int) error {
	if n != s.Len {
		return errors.New("短信参数错误")
	}
	return nil
}

// Parse 解析处理
func (s AliTpl) Parse(data ...interface{}) string {
	// 解析模板 模板 验证对应参数是否对照
	return fmt.Sprintf(s.Params, data...)
}

type aliReturn struct {
	RequestID string `json:"RequestId"` // 请求ID
	Code      string // 状态码
	Message   string // 状态码描述
	BizID     string `json:"BizId"` // 发送回执ID
}

// aliCodes ...
var aliCodes = map[string]string{
	"OK":                              "请求成功",
	"isp.RAM_PERMISSION_DENY":         "RAM权限DENY",
	"isv.OUT_OF_SERVICE":              "业务停机",
	"isv.PRODUCT_UN_SUBSCRIPT":        "未开通云通信产品的阿里云客户",
	"isv.PRODUCT_UNSUBSCRIBE":         "产品未开通",
	"isv.ACCOUNT_NOT_EXISTS":          "账户不存在",
	"isv.ACCOUNT_ABNORMAL":            "账户异常",
	"isv.SMS_TEMPLATE_ILLEGAL":        "短信模板不合法",
	"isv.SMS_SIGNATURE_ILLEGAL":       "短信签名不合法",
	"isv.INVALID_PARAMETERS":          "参数异常",
	"isp.SYSTEM_ERROR":                "系统错误",
	"isv.MOBILE_NUMBER_ILLEGAL":       "非法手机号",
	"isv.MOBILE_COUNT_OVER_LIMIT":     "手机号码数量超过限制",
	"isv.TEMPLATE_MISSING_PARAMETERS": "模板缺少变量",
	"isv.BUSINESS_LIMIT_CONTROL":      "业务限流",
	"isv.INVALID_JSON_PARAM":          "JSON参数不合法，只接受字符串值",
	"isv.BLACK_KEY_CONTROL_LIMIT":     "黑名单管控",
	"isv.PARAM_LENGTH_LIMIT":          "参数超出长度限制",
	"isv.PARAM_NOT_SUPPORT_URL":       "不支持URL",
	"isv.AMOUNT_NOT_ENOUGH":           "账户余额不足",
}

// Tpl 模板
func (s AliSender) Tpl(smsType uint) (SmsTpl, error) {
	for _, v := range s.Tpls {
		if v.Type == smsType {
			return v, nil
		}
	}

	return nil, errors.New("短信模板不存在，请更换短信类型！")
}

// Send 仅仅发送
func (s AliSender) Send(smsType uint, tel string, data ...interface{}) error {

	tpli, err := s.Tpl(smsType)
	if err != nil {
		return err
	}
	tpl, _ := tpli.(AliTpl)

	code := tpl.Parse(data...)

	var params = map[string]string{}
	t := time.Now().In(BeijingLocation).UTC().Format("2006-01-02T15:04:05Z")
	var nonce = strings.Replace(uuid.NewV4().String(), "-", "", 4)

	// 1. 系统参数
	params["SignatureMethod"] = "HMAC-SHA1"
	params["SignatureNonce"] = nonce
	params["AccessKeyId"] = s.Key
	params["SignatureVersion"] = "1.0"
	params["Timestamp"] = t
	params["Format"] = "JSON"

	// 2. 业务参数
	params["Action"] = "SendSms"
	params["Version"] = "2017-05-25"
	params["RegionId"] = "cn-hangzhou"

	params["PhoneNumbers"] = tel
	params["SignName"] = s.SignName
	params["TemplateCode"] = tpl.ID
	// `{"customer":"test"}`
	params["TemplateParam"] = code
	// params["OutId"] = ""

	// 3. 去除签名关键字Key
	delete(params, "Signature")

	// 4. 参数KEY排序
	parasIndex := make([]string, 0)
	for k := range params {
		parasIndex = append(parasIndex, k)
	}
	sort.Strings(parasIndex)

	// 5. 构造待签名的字符串
	sortedQueryString := ""
	for _, v := range parasIndex {
		sortedQueryString = sortedQueryString + "&" + specialUrlEncode(v) + "=" + specialUrlEncode(params[v])
	}

	// 去除第一个多余的&符号
	sortedQueryString = sortedQueryString[1:]

	stringToSign := "GET" + "&" + specialUrlEncode("/") + "&" + specialUrlEncode(sortedQueryString)

	signStr := sign(s.Secret+"&", stringToSign)

	// 6. 签名最后也要做特殊URL编码
	signature := specialUrlEncode(signStr)

	// 最终打印出合法GET请求的URL
	urlStr := "http://dysmsapi.aliyuncs.com/?Signature=" + signature + "&" + sortedQueryString

	// fmt.Println(urlStr)

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
	re := &aliReturn{}
	if err := json.Unmarshal(body, re); err == nil {
		if re.Code == "OK" {
			return nil
		} else {
			return errors.New(re.Message)
		}
	}

	return nil
}

type aliRequest struct {
	PhoneNumbers    string // 手机号
	SignName        string // 签名
	TemplateCode    string // 模板标识
	TemplateParsm   string // 模板参数
	SmsUpExtendCode string // 上行短信扩展吗
	OutID           string `json:"OutId"` // 外部流水字段
}

func randomString(length int) string {
	base := "abcdefghijklmnopqrstuvwxyz1234567890"
	result := ""
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(result) < length {
		index := r.Intn(len(base))
		result = result + base[index:index+1]
	}
	return result
}

func specialUrlEncode(value string) string {
	result := url.QueryEscape(value)
	result = strings.Replace(result, "+", "%20", -1)
	result = strings.Replace(result, "*", "%2A", -1)
	result = strings.Replace(result, "%7E", "~", -1)
	return result
}

func sign(accessSecret, strToSign string) string {
	mac := hmac.New(sha1.New, []byte(accessSecret))
	mac.Write([]byte(strToSign))
	signData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signData)
}
