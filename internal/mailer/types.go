package mailer

type ResetPasswordData struct {
	ToAddress string
	Username  string
	Token     string
}

type ContactFormData struct {
	UserName  string
	UserEmail string
	Subject   string
	Message   string

	Error  string
	Border string
}
