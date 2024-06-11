package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddr = "smtp.163.com"
	smtpServerAddr = "smtp.163.com:25"
)

type EmailSender interface {
	SendMail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type NetEaseMailSender struct {
	name string
	fromMailAddress string
	fromMailPassword string
}

func NewNetEaseMailSender (name, fromMailAddress, fromMailPassword string) EmailSender {
	return &NetEaseMailSender{ 
		name: name,
		fromMailAddress: fromMailAddress,
		fromMailPassword: fromMailPassword,
	}
}

func (sender *NetEaseMailSender) SendMail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	// create a new email
	e := &email.Email {
		To: to,
		From: fmt.Sprintf("%s <%s>", sender.name, sender.fromMailAddress), 
		Subject: subject, 
		// Text: []byte("Text Body is, of course, supported!"),
		HTML: []byte(content), 
		// Headers: textproto.MIMEHeader{},
	}

	for _, f := range attachFiles {
		if _, err := e.AttachFile(f); err != nil {
			return fmt.Errorf("fail to attach file %s: %w", f, err)
		}
	}

	// Set up authentication information.
	auth := smtp.PlainAuth("", sender.fromMailAddress, sender.fromMailPassword, smtpAuthAddr)
	// send email
	return e.Send(smtpServerAddr, auth)
}