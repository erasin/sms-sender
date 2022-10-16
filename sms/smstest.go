package sms

// SmsTestSender 测试用
type SmsTestSender struct{}

// Send 发送短信
func (s SmsTestSender) Send(smsType uint, tel string, data ...interface{}) error {
	return nil
}

// Tpl 模板
func (s SmsTestSender) Tpl(smsType uint) (SmsTpl, error) {
	return nil, nil
}
