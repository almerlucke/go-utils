// Package email defines the interface for an email send service, loosely based on AWS SES structure
package email

// Destination of the email
type Destination struct {
	BccAddresses []string
	CcAddresses  []string
	ToAddresses  []string
}

// Message of the email
type Message struct {
	Body    *Body
	Subject *Content
}

// Body of the email
type Body struct {
	HTML *Content
	Text *Content
}

// Content of an email part
//
// By default, the text must be 7-bit ASCII, due to the constraints of the SMTP
// protocol. If the text must contain any other characters, then you must also
// specify a character set. Examples include UTF-8, ISO-8859-1, and Shift_JIS
type Content struct {
	Charset string
	Data    string
}

// SendEmailInput input for sending the email
type SendEmailInput struct {
	Destination      *Destination
	Message          *Message
	ReplyToAddresses []string
	ReturnPath       string
	Source           string
}

// SendRawEmailInput input for sending raw email
type SendRawEmailInput struct {
	RawMessage []byte
}

// Mailer interface
type Mailer interface {
	SendEmail(*SendEmailInput) error
	SendRawEmail(*SendRawEmailInput) error
}
