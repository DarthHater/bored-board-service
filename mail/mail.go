/* Copyright 2017 Jeffry Hesse

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */
package mail

import (
	"github.com/DarthHater/bored-board-service/constants"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

func init() {
	setupViper()
}

func setupViper() {
	log.WithFields(log.Fields{
		"package": "mail",
	}).Debug("Setting up Viper")

	viper.SetEnvPrefix(constants.EnvironmentVariablePrefix)
	viper.BindEnv(constants.RegistrationEmailEnvVariable)
	viper.BindEnv(constants.SendGridAPIKeyEnvVariable)
	viper.BindEnv(constants.EmailFromAddressEnvVariable)
	viper.BindEnv(constants.EmailFromNameEnvVariable)
}

// SendNewUserEmail function to send a new user registration email
func SendNewUserEmail(recipient string, subject string, userName string, boardURLVerify string, donateURL string) {
	m := mail.NewV3Mail()
	m.Subject = subject
	p := mail.NewPersonalization()

	to := mail.NewEmail(userName, recipient)
	log.WithFields(log.Fields{
		"emailFromName":    viper.GetString(constants.EmailFromNameEnvVariable),
		"emailFromAddress": viper.GetString(constants.EmailFromAddressEnvVariable),
	}).Info("Attempting to send New User email")
	p.AddTos(to)

	from := mail.NewEmail(
		viper.GetString(constants.EmailFromNameEnvVariable),
		viper.GetString(constants.EmailFromAddressEnvVariable),
	)
	m.SetFrom(from)

	m.SetTemplateID(viper.GetString(constants.RegistrationEmailEnvVariable))

	t := time.Now()

	p.SetDynamicTemplateData("userName", userName)
	p.SetDynamicTemplateData("boardURLVerify", boardURLVerify)
	p.SetDynamicTemplateData("donateURL", donateURL)
	p.SetDynamicTemplateData("year", strconv.Itoa(t.Year()))

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(
		viper.GetString(constants.SendGridAPIKeyEnvVariable),
		constants.SendgridSendMailAPIPathV3,
		constants.SendGridAPIBasePath,
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
