package mailer

type ResetPasswordData struct {
	ToAddress string
	Username  string
	Token     string
}
