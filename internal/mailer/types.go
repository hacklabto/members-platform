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

type NewApplicantData struct {
	ToAddresses string

	Name                   string
	Nickname               string
	ContactEmail           string
	Website                string
	ProfileURL             string
	Sponsor                string
	SecondSponsor          string
	Twitter                string
	Facebook               string
	BioReason              string
	HeardFrom              string
	PreferredGenderPronoun string
}
