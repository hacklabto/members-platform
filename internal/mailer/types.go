package mailer

type ResetPasswordData struct {
	ToAddress string
	Username  string
	Token     string
}

type ListsErrorReply struct {
	ToAddress string
	Errors    string
}
