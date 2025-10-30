package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

func DoSendEmail(email, content string) error {
	return DoSendEmailInner("operations+automated@hacklab.to", email, content)
}

func DoSendEmailInner(src, email, content string) error {
	smtpServer := os.Getenv("SMTP_URL")
	if smtpServer == "" {
		return fmt.Errorf("missing SMTP_URL in environment")
	}

	conn, err := smtp.Dial(smtpServer)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer conn.Close()

	if hello := os.Getenv("SMTP_HELLO"); hello != "" {
		log.Printf("sending hello: %v\n", hello)
		if err := conn.Hello(hello); err != nil {
			return fmt.Errorf("conn.Hello: %w", err)
		}
	}

	if os.Getenv("SMTP_USE_STARTTLS") == "true" {
		log.Printf("starting tls\n")
		if err := conn.StartTLS(&tls.Config{
			ServerName: strings.Split(smtpServer, ":")[0],
		}); err != nil {
			return fmt.Errorf("conn.StartTLS: %w", err)
		}
	}

	if err := conn.Mail(src); err != nil {
		return fmt.Errorf("conn.Mail: %w", err)
	}

	if err := conn.Rcpt(email); err != nil {
		return fmt.Errorf("conn.Rcpt: %w", err)
	}

	wc, err := conn.Data()
	if err != nil {
		return fmt.Errorf("conn.Data: %w", err)
	}
	defer wc.Close()

	if _, err := wc.Write([]byte(content)); err != nil {
		return fmt.Errorf("write email content: %w", err)
	}

	return nil
}
