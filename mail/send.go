package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
)

type EmailClient interface {
	Send(ctx context.Context, req *SendEmailRequest) error
}

type SendEmailRequest struct {
	To        string
	From      string
	Subject   string
	PlainText string
}

type smtpClient struct {
	host     string
	port     string
	username string
	password string
}

func NewSMTPClient(host, port, username, password string) EmailClient {
	return &smtpClient{
		host, port, username, password,
	}
}

// Uses env var SMTP settings to establish new mail client. If settings are missing, nil is returned.
// Requires KEEL_SMTP_HOST, KEEL_SMTP_PORT, KEEL_SMTP_USER, and KEEL_SMTP_PASSWORD.
func NewSMTPClientFromEnv() EmailClient {
	var host, port, username, password string

	if host = os.Getenv("KEEL_SMTP_HOST"); host == "" {
		return nil
	}
	if port = os.Getenv("KEEL_SMTP_PORT"); port == "" {
		return nil
	}
	if username = os.Getenv("KEEL_SMTP_USER"); username == "" {
		return nil
	}
	if password = os.Getenv("KEEL_SMTP_PASSWORD"); password == "" {
		return nil
	}

	return &smtpClient{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (c *smtpClient) Send(ctx context.Context, req *SendEmailRequest) error {
	host := fmt.Sprintf("%s:%s", c.host, c.port)
	auth := smtp.PlainAuth("", c.username, c.password, c.host)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s.\r\n",
		req.From, req.To, req.Subject, req.PlainText)

	return smtp.SendMail(host, auth, req.From, []string{req.To}, []byte(msg))
}

type noOpClient struct {
}

// No op email client does not send a mail.
func NoOpClient() EmailClient {
	return &noOpClient{}
}

func (c *noOpClient) Send(context.Context, *SendEmailRequest) error {
	return nil
}
