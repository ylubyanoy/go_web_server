package services

// MailData represents the data to be sent to the new user.
type MailData struct {
	Username string
	Code     string
}
