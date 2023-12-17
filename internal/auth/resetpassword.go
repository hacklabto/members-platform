package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"members-platform/internal/mailer"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// poor hacker's jwt
func CreateResetToken(username string) string {
	signkey := os.Getenv("PASSWD_RESET_HASHER_SECRET")
	if signkey == "" {
		panic(fmt.Errorf("missing PASSWD_RESET_HASHER_SECRET in environment"))
	}

	hmacer := hmac.New(sha256.New, []byte(signkey))

	var b strings.Builder

	b.WriteString(base64.RawURLEncoding.EncodeToString([]byte(username)))
	b.WriteString(".")
	b.WriteString(base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(int(time.Now().UTC().Unix())))))

	hmacer.Write([]byte(b.String()))
	b.WriteString(".")
	b.WriteString(base64.RawURLEncoding.EncodeToString(hmacer.Sum(nil)))

	return b.String()
}

// todo: validate reset token

func SendResetEmail(email, username, token string) error {
	d := mailer.ResetPasswordData{
		ToAddress: email,
		Username:  username,
		Token:     token,
	}

	text, err := mailer.ExecuteTemplate("reset-password", d)
	if err != nil {
		return fmt.Errorf("build email body: %w", err)
	}

	smtpServer := os.Getenv("SMTP_SERVER")
	if smtpServer == "" {
		return fmt.Errorf("missing SMTP_SERVER in environment")
	}

	conn, err := smtp.Dial(smtpServer)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer conn.Close()

	if err := conn.Mail("operations+automated@hacklab.to"); err != nil {
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

	if _, err := wc.Write([]byte(text)); err != nil {
		return fmt.Errorf("write email body: %w", err)
	}

	return nil
}
