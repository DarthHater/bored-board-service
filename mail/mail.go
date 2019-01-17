package mail

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

const (
	sendgridSendMailAPIPathV3    string = "/v3/mail/send"
	sendGridAPIBasePath          string = "https://api.sendgrid.com"
	registrationEmailEnvVariable string = "REGISTRATION_EMAIL_TEMPLATE_ID"
	sendGridAPIKeyEnvVariable    string = "SENDGRID_API_KEY"
	emailFromAddressEnvVariable  string = "EMAIL_FROM_ADDRESS"
	emailFromNameEnvVariable     string = "EMAIL_FROM_NAME"
)

func init() {
	setupViper()
}

func setupViper() {
	log.WithFields(log.Fields{
		"package": "mail",
	}).Debug("Setting up Viper")

	viper.SetEnvPrefix("BBS")
	viper.BindEnv(registrationEmailEnvVariable)
	viper.BindEnv(sendGridAPIKeyEnvVariable)
	viper.BindEnv(emailFromAddressEnvVariable)
	viper.BindEnv(emailFromNameEnvVariable)
}

// SendNewUserEmail function to send a new user registration email
func SendNewUserEmail(recipient string, subject string, userName string, boardURLVerify string, donateURL string) {
	m := mail.NewV3Mail()
	m.Subject = subject
	p := mail.NewPersonalization()

	to := mail.NewEmail(userName, recipient)
	log.WithFields(log.Fields{
		"emailFromName":    viper.GetString(emailFromNameEnvVariable),
		"emailFromAddress": viper.GetString(emailFromAddressEnvVariable),
	}).Info("Attempting to send New User email")
	p.AddTos(to)

	from := mail.NewEmail(
		viper.GetString(emailFromNameEnvVariable),
		viper.GetString(emailFromAddressEnvVariable),
	)
	m.SetFrom(from)

	m.SetTemplateID(viper.GetString(registrationEmailEnvVariable))

	t := time.Now()

	p.SetDynamicTemplateData("userName", userName)
	p.SetDynamicTemplateData("boardURLVerify", boardURLVerify)
	p.SetDynamicTemplateData("donateURL", donateURL)
	p.SetDynamicTemplateData("year", strconv.Itoa(t.Year()))

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(
		viper.GetString(sendGridAPIKeyEnvVariable),
		sendgridSendMailAPIPathV3,
		sendGridAPIBasePath,
	)
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)

	response, err := sendgrid.API(request)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"userName":  userName,
			"recipient": recipient,
		}).Error("Error sending New User Registration email")
	} else {
		log.WithFields(log.Fields{
			"responseCode": response.StatusCode,
			"userName":     userName,
		}).Info("New User Registration Email sent successfully")
	}
}
