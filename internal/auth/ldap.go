package auth

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/go-ldap/ldap"
)

func GetEmailFromUsername(bindDN, bindPassword, targetUsername string) (string, error) {
	ldapURL := os.Getenv("LDAP_URL")
	if ldapURL == "" {
		return "", fmt.Errorf("missing LDAP_URL in environment")
	}
	conn, err := ldap.DialTLS("tcp", ldapURL, &tls.Config{
		// todo(infra): don't
		InsecureSkipVerify: true,
	})
	if err != nil {
		return "", fmt.Errorf("dial ldap: %w", err)
	}
	defer conn.Close()

	if err := conn.Bind(bindDN, bindPassword); err != nil {
		return "", fmt.Errorf("bind ldap: %w", err)
	}

	searchRequest := ldap.NewSearchRequest(
		"ou=people,dc=hacklab,dc=to",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(uid=%s))", targetUsername),
		[]string{"mail"},
		nil,
	)

	res, err := conn.Search(searchRequest)
	if err != nil {
		return "", err
	}

	for _, entry := range res.Entries {
		return entry.GetAttributeValue("mail"), nil
	}

	return "", fmt.Errorf("user not found")
}

func AuthenticateUser(username, password string) (bool, error) {
	ldapURL := os.Getenv("LDAP_URL")
	if ldapURL == "" {
		return false, fmt.Errorf("missing LDAP_URL in environment")
	}
	conn, err := ldap.DialTLS("tcp", ldapURL, &tls.Config{
		// todo(infra): don't
		InsecureSkipVerify: true,
	})
	if err != nil {
		return false, fmt.Errorf("dial ldap: %w", err)
	}
	defer conn.Close()

	if err := conn.Bind(fmt.Sprintf("uid=%s,ou=people,dc=hacklab,dc=to", username), password); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return false, nil
		}
		return false, fmt.Errorf("bind ldap: %w", err)
	}

	return true, nil
}

// todo: authenticate service

// DoChangePassword can be used for both password reset and password change
// if password reset, bind with admin user
// otherwise you can bind with your current credentials to change your account password
func DoChangePassword(bindDN, bindPassword, targetDN, newPassword string) error {
	ldapURL := os.Getenv("LDAP_URL")
	if ldapURL == "" {
		return fmt.Errorf("missing LDAP_URL in environment")
	}
	conn, err := ldap.DialTLS("tcp", ldapURL, &tls.Config{
		// todo(infra): don't
		InsecureSkipVerify: true,
	})
	if err != nil {
		return fmt.Errorf("dial ldap: %w", err)
	}
	defer conn.Close()

	// bind as admin user
	if err := conn.Bind(bindDN, bindPassword); err != nil {
		return fmt.Errorf("bind ldap: %w", err)
	}

	oldPassword := ""
	if bindDN == targetDN {
		oldPassword = bindPassword
	}

	_, err = conn.PasswordModify(ldap.NewPasswordModifyRequest(targetDN, oldPassword, newPassword))
	return err
}
