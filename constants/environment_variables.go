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
package constants

const (
	EnvironmentVariablePrefix    string = "BBS"
	RegistrationEmailEnvVariable string = "REGISTRATION_EMAIL_TEMPLATE_ID"
	SendGridAPIKeyEnvVariable    string = "SENDGRID_API_KEY"
	EmailFromAddressEnvVariable  string = "EMAIL_FROM_ADDRESS"
	EmailFromNameEnvVariable     string = "EMAIL_FROM_NAME"
	BoardURLVerifyEnvVariable    string = "BOARD_URL_VERIFY"
	BoardURLDonateEnvVariable    string = "BOARD_URL_DONATE"
	BoardURLCorsEnvVariable      string = "BOARD_URL_CORS"
	BoardSendNewUserEmailSubject string = "BOARD_SEND_NEW_USER_EMAIL_SUBJECT"
	RedisURLEnvVariable          string = "REDIS_URL"
)
