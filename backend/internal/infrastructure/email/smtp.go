package email

import (
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTP struct {
	address string
	from    string
}

func NewSMTP(host, port, from string) (*SMTP, error) {
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("SMTP_PORT must be a valid TCP port")
	}
	if strings.TrimSpace(host) == "" {
		return nil, fmt.Errorf("SMTP_HOST must not be empty")
	}
	if strings.TrimSpace(from) == "" || strings.ContainsAny(from, "\r\n<>") {
		return nil, fmt.Errorf("SMTP_FROM must be a valid email address")
	}

	return &SMTP{
		address: net.JoinHostPort(host, port),
		from:    from,
	}, nil
}

func (s *SMTP) SendVerificationCode(to, code string) error {
	return s.sendCode(to, code, "アカウント確認コード", "確認", "アカウント登録の確認コード")
}

func (s *SMTP) SendAccountRecoveryCode(to, code string) error {
	return s.sendCode(to, code, "アカウント復旧コード", "復旧確認", "アカウント復旧の確認コード")
}

func (s *SMTP) sendCode(to, code, subject, purpose, description string) error {
	message := []byte(
		"To: " + to + "\r\n" +
			"From: " + s.from + "\r\n" +
			"Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			description + "をお送りします。\r\n" +
			"\r\n" +
			"確認コード: " + code + "\r\n" +
			"\r\n" +
			"この" + purpose + "コードの有効期限は10分です。\r\n",
	)
	connection, err := net.DialTimeout("tcp", s.address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("set SMTP deadline: %w", err)
	}

	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		return fmt.Errorf("parse SMTP server address: %w", err)
	}
	client, err := smtp.NewClient(connection, host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}
