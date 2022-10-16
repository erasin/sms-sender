// Package sms 短信发送驱动
package sms

import "time"

// SmsSender 短信接口
type SmsSender interface {
	// Send 发送
	// @params smsType 发送类型，关联模板
	// @params tel 发送的手机号
	// @params code 发送内容
	Send(smsType uint, tel string, data ...interface{}) error
	// Tpl 获取模板
	Tpl(smsType uint) (tpl SmsTpl, err error)
}

type SmsTpl interface {
	ValidLen(n int) error
	Parse(data ...interface{}) string
}

// BeijingLocation 北京时间
var BeijingLocation = time.FixedZone("Asia/Shanghai", 3600<<3) // 3600 * 8
