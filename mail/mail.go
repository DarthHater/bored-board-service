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
	SENDGRID_SEND_MAIL_API_PATH_V3  string = "/v3/mail/send"
	SENDGRID_API_BASE_PATH          string = "https://api.sendgrid.com"
	REGISTRATION_EMAIL_ENV_VARIABLE string = "REGISTRATION_EMAIL_TEMPLATE_ID"
	SENDGRID_API_KEY_ENV_VARIABLE   string = "SENDGRID_API_KEY"
	EMAIL_FROM_ADDRESS_ENV_VARIABLE string = "EMAIL_FROM_ADDRESS"
	EMAIL_FROM_NAME_ENV_VARIABLE    string = "EMAIL_FROM_NAME"
)

func init() {
	setupViper()
}

func setupViper() {
	log.WithFields(log.Fields{
		"package": "mail",
	}).Debug("Setting up Viper")

	viper.SetEnvPrefix("BBS")
	viper.BindEnv(REGISTRATION_EMAIL_ENV_VARIABLE)
	viper.BindEnv(SENDGRID_API_KEY_ENV_VARIABLE)
	viper.BindEnv(EMAIL_FROM_ADDRESS_ENV_VARIABLE)
	viper.BindEnv(EMAIL_FROM_NAME_ENV_VARIABLE)
}

func SendEmail(recipient string, subject string, userName string, boardUrlVerify string, donateUrl string) {
	m := mail.NewV3Mail()
	m.Subject = subject
	p := mail.NewPersonalization()

	to := mail.NewEmail(userName, recipient)
	log.WithFields(log.Fields{
		"emailFromName":    viper.GetString(EMAIL_FROM_NAME_ENV_VARIABLE),
		"emailFromAddress": viper.GetString(EMAIL_FROM_ADDRESS_ENV_VARIABLE),
	}).Info("Attempting to send New User email")
	p.AddTos(to)

	from := mail.NewEmail(
		viper.GetString(EMAIL_FROM_NAME_ENV_VARIABLE),
		viper.GetString(EMAIL_FROM_ADDRESS_ENV_VARIABLE),
	)
	m.SetFrom(from)

	m.SetTemplateID(viper.GetString(REGISTRATION_EMAIL_ENV_VARIABLE))

	t := time.Now()

	p.SetDynamicTemplateData("userName", userName)
	p.SetDynamicTemplateData("boardUrlVerify", boardUrlVerify)
	p.SetDynamicTemplateData("donateUrl", donateUrl)
	p.SetDynamicTemplateData("year", strconv.Itoa(t.Year()))

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(
		viper.GetString(SENDGRID_API_KEY_ENV_VARIABLE),
		SENDGRID_SEND_MAIL_API_PATH_V3,
		SENDGRID_API_BASE_PATH,
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
