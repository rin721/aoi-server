package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type SMTPNotifierConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	StartTLS bool
}

type SMTPNotifier struct {
	cfg SMTPNotifierConfig
}

func NewSMTPNotifier(cfg SMTPNotifierConfig) (*SMTPNotifier, error) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.From = strings.TrimSpace(cfg.From)
	cfg.FromName = strings.TrimSpace(cfg.FromName)
	if cfg.Port <= 0 {
		cfg.Port = 587
	}
	if cfg.Host == "" || cfg.From == "" {
		return nil, ErrInvalidInput
	}
	return &SMTPNotifier{cfg: cfg}, nil
}

func (n *SMTPNotifier) SendInvitation(ctx context.Context, notice InvitationNotice) error {
	return n.send(ctx, notice.Email, "Aoi Admin invitation", fmt.Sprintf("You have been invited to Aoi Admin.\r\n\r\nOpen this link to accept the invitation:\r\n%s\r\n", notice.URL))
}

func (n *SMTPNotifier) SendPasswordReset(ctx context.Context, notice PasswordResetNotice) error {
	return n.send(ctx, notice.Email, "Aoi Admin password reset", fmt.Sprintf("A password reset was requested for your Aoi Admin account.\r\n\r\nOpen this link to reset your password:\r\n%s\r\n", notice.URL))
}

func (n *SMTPNotifier) send(ctx context.Context, to string, subject string, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	to = strings.TrimSpace(to)
	if to == "" {
		return ErrInvalidInput
	}
	from := mail.Address{Name: n.cfg.FromName, Address: n.cfg.From}
	addr := net.JoinHostPort(n.cfg.Host, fmt.Sprintf("%d", n.cfg.Port))
	message := []byte(strings.Join([]string{
		"From: " + from.String(),
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n"))

	if !n.cfg.StartTLS {
		return smtp.SendMail(addr, n.auth(), n.cfg.From, []string{to}, message)
	}
	return n.sendStartTLS(addr, to, message)
}

func (n *SMTPNotifier) sendStartTLS(addr string, to string, message []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: n.cfg.Host, MinVersion: tls.VersionTLS12}
	if err := client.StartTLS(tlsCfg); err != nil {
		return err
	}
	if auth := n.auth(); auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(n.cfg.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (n *SMTPNotifier) auth() smtp.Auth {
	if n.cfg.Username == "" {
		return nil
	}
	return smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.Host)
}
