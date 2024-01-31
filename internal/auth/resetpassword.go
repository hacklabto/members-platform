package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"members-platform/internal/db"
	"members-platform/internal/mailer"
	"time"
)

func CreateResetToken(username string) (string, error) {
	b := make([]byte, 48)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	return token, db.Redis.Set(context.Background(), "reset-token:"+token, username, time.Hour*24).Err()
}

func ValidateResetToken(token string) (string, bool) {
	v, err := db.Redis.Get(context.Background(), "reset-token:"+token).Result()
	switch {
	case err != nil:
		log.Println(err)
		return "", false
	}
	return v, true
}

func SendResetEmail(email, username, token string) error {
	d := mailer.ResetPasswordData{
		ToAddress: email,
		Username:  username,
		Token:     token,
	}

	content, err := mailer.ExecuteTemplate("reset-password", d)
	if err != nil {
		return fmt.Errorf("build email content: %w", err)
	}

	return mailer.DoSendEmail(email, content)
}

func SendResetRestrictedEmail(email, username string) error {
	d := mailer.ResetPasswordData{
		ToAddress: email,
		Username:  username,
	}

	content, err := mailer.ExecuteTemplate("reset-password-restricted", d)
	if err != nil {
		return fmt.Errorf("build email content: %w", err)
	}

	return mailer.DoSendEmail(email, content)
}

func SendConfirmationEmail(email, username string) error {
	d := mailer.ResetPasswordData{
		ToAddress: email,
		Username:  username,
	}

	content, err := mailer.ExecuteTemplate("reset-password-confirm", d)
	if err != nil {
		return fmt.Errorf("build email content: %w", err)
	}

	return mailer.DoSendEmail(email, content)
}
