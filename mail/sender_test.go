package mail

import (
	"testing"

	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {
	// skip test send email
	if testing.Short() {
		t.Skip()
	}

	config, err := utils.LoadConfig("../")
	require.NoError(t, err)


	netEaseMail := NewNetEaseMailSender(
		config.EmailSenderName, 
		config.EmailSenderAddress, 
		config.EmailSenderPassword,
	)		

	subject := "A Test Email"
	content := `
	<h1> Hello World </h1>
	<p> This is a test mail </p>
	`
	to := []string{"mail-address@example.com"}
	err = netEaseMail.SendMail(subject, content, to, nil, nil, nil)
	require.NoError(t, err)
}