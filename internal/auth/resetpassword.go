package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"members-platform/internal/db"
	"members-platform/internal/mailer"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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

func ValidateResetToken(token string) (string, bool) {
	signkey := os.Getenv("PASSWD_RESET_HASHER_SECRET")
	if signkey == "" {
		panic(fmt.Errorf("missing PASSWD_RESET_HASHER_SECRET in environment"))
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}

	username, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}

	timestamp, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	if t, err := strconv.Atoi(string(timestamp)); err != nil {
		return "", false
	} else {
		// if token created > 1 day ago
		if (int(time.Now().UTC().Unix()) - t) > int((time.Hour * 24).Seconds()) {
			return "", false
		}
	}

	hmac_from_user, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", false
	}

	hmacer := hmac.New(sha256.New, []byte(signkey))

	if _, err := hmacer.Write([]byte(parts[0] + "." + parts[1])); err != nil {
		return "", false
	}

	if !hmac.Equal(hmac_from_user, hmacer.Sum(nil)) {
		return "", false
	}

	_, err = db.RedisDB.Get(context.Background(), "used-reset-token:"+token).Result()
	switch {
	// token not used
	case err == redis.Nil:
		return string(username), true
	case err != nil:
		log.Println(err)
		return "", false
	}
	// if err == nil, token already used
	return "", false
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
