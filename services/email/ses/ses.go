package ses

import (
	"github.com/almerlucke/go-utils/services/email"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// Mailer wrapper around SES
type Mailer struct {
	ses *ses.SES
}

// New AWS SES wrapper for emailer interface
func New(session *session.Session) *Mailer {
	return &Mailer{
		ses: ses.New(session),
	}
}

func contentToAWSEmailContent(content *email.Content) *ses.Content {
	var charset *string = nil

	if content.Charset != "" {
		charset = aws.String(content.Charset)
	}

	return &ses.Content{
		Charset: charset,
		Data:    aws.String(content.Data),
	}
}

func bodyToAWSEmailBody(body *email.Body) *ses.Body {
	b := &ses.Body{}

	if body.HTML != nil {
		b.Html = contentToAWSEmailContent(body.HTML)
	}

	if body.Text != nil {
		b.Text = contentToAWSEmailContent(body.Text)
	}

	return b
}

func messageToAWSEmailMessage(message *email.Message) *ses.Message {
	m := &ses.Message{}

	if message.Body != nil {
		m.Body = bodyToAWSEmailBody(message.Body)
	}

	if message.Subject != nil {
		m.Subject = contentToAWSEmailContent(message.Subject)
	}

	return m
}

func stringSliceToAWSStringSlice(s []string) []*string {
	as := make([]*string, len(s))

	for i, v := range s {
		as[i] = aws.String(v)
	}

	return as
}

func destinationToAWSEmailDestination(destination *email.Destination) *ses.Destination {
	d := &ses.Destination{}

	if destination.BccAddresses != nil {
		d.BccAddresses = stringSliceToAWSStringSlice(destination.BccAddresses)
	}

	if destination.CcAddresses != nil {
		d.CcAddresses = stringSliceToAWSStringSlice(destination.CcAddresses)
	}

	if destination.ToAddresses != nil {
		d.ToAddresses = stringSliceToAWSStringSlice(destination.ToAddresses)
	}

	return d
}

func sendEmailInputToAWSSendEmailInput(input *email.SendEmailInput) *ses.SendEmailInput {
	i := &ses.SendEmailInput{}

	if input.Destination != nil {
		i.Destination = destinationToAWSEmailDestination(input.Destination)
	}

	if input.Message != nil {
		i.Message = messageToAWSEmailMessage(input.Message)
	}

	if input.ReplyToAddresses != nil {
		i.ReplyToAddresses = stringSliceToAWSStringSlice(input.ReplyToAddresses)
	}

	if input.ReturnPath != "" {
		i.ReturnPath = aws.String(input.ReturnPath)
	}

	if input.Source != "" {
		i.Source = aws.String(input.Source)
	}

	return i
}

// SendEmail send email
func (email *Mailer) SendEmail(input *email.SendEmailInput) error {
	_, err := email.ses.SendEmail(sendEmailInputToAWSSendEmailInput(input))
	return err
}

// SendRawEmail send raw email
func (email *Mailer) SendRawEmail(input *email.SendRawEmailInput) error {
	_, err := email.ses.SendRawEmail(&ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: input.RawMessage,
		},
	})
	return err
}
