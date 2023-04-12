package mail

import (
	"fmt"
	"net/smtp"
)

type Client struct {
	Enabled  bool
	ConnInfo *SmtpConnInfo
}

type SmtpConnInfo struct {
	Host     string
	Port     int
	From     string
	Username string
	Password string
}

func (c Client) SendMail(to []string, message string) error {
	if !c.Enabled {
		return nil
	}

	host := fmt.Sprintf("%s:%v", c.ConnInfo.Host, c.ConnInfo.Port)
	auth := smtp.PlainAuth("", c.ConnInfo.Username, c.ConnInfo.Password, c.ConnInfo.Host)
	contents := []byte(message)

	// TODO: run on a new thread?
	err := smtp.SendMail(host, auth, c.ConnInfo.From, to, contents)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) SendResetPasswordMail(to string, redirectUrl string) error {
	// TODO: make a pretty email
	return c.SendMail([]string{to}, redirectUrl)
}
