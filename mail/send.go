package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/db")

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

func Local() *Client {
	var host, from, username, password string
	var port int
	var err error

	if port, err = strconv.Atoi(os.Getenv("KEEL_SMTP_PORT")); err != nil {
		panic("set KEEL_SMTP_PORT to enable SMTP")
	}

	if host = os.Getenv("KEEL_SMTP_HOST"); host == "" {
		panic("set KEEL_SMTP_HOST to enable SMTP")
	}

	if from = os.Getenv("KEEL_SMTP_FROM"); from == "" {
		panic("set KEEL_SMTP_FROM to enable SMTP")
	}

	if username = os.Getenv("KEEL_SMTP_USER"); username == "" {
		panic("set KEEL_SMTP_USER to enable SMTP")
	}

	if password = os.Getenv("KEEL_SMTP_PASSWORD"); password == "" {
		panic("set KEEL_SMTP_PASSWORD to enable SMTP")
	}

	smtpConnInfo := &SmtpConnInfo{
		Host:     host,
		Port:     port,
		From:     from,
		Username: username,
		Password: password,
	}

	return &Client{
		Enabled:  true,
		ConnInfo: smtpConnInfo,
	}
}

func Disabled() *Client {
	return &Client{
		Enabled:  false,
		ConnInfo: nil,
	}
}

func (c Client) SendMail(ctx context.Context, to []string, subject string, contents string) {
	if !c.Enabled {
		return
	}

	_, span := tracer.Start(ctx, "SendMail")

	host := fmt.Sprintf("%s:%v", c.ConnInfo.Host, c.ConnInfo.Port)
	auth := smtp.PlainAuth("", c.ConnInfo.Username, c.ConnInfo.Password, c.ConnInfo.Host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s.\r\n",
		c.ConnInfo.From, strings.Join(to, ","), subject, contents)

	go func() {
		err := smtp.SendMail(host, auth, c.ConnInfo.From, to, []byte(msg))
		if err != nil {
			span.SetStatus(codes.Error, "Email failured to send")
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "Email successfully sent")
		}
		span.End()
	}()
}

func (c Client) SendResetPasswordMail(ctx context.Context, to string, redirectUrl string) {
	subject := "[Keel] Reset password request"
	contents := fmt.Sprintf("Please follow this link to reset your password: %s", redirectUrl)
	c.SendMail(ctx, []string{to}, subject, contents)
}
