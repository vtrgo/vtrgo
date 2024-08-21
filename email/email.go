package email

import (
	"encoding/json"
	"html/template"
	"os"
	"strings"

	"gopkg.in/gomail.v2"
)

type Config struct {
	SMTPHost string `json:"SMTPHost"`
	SMTPPort int    `json:"SMTPPort"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type EmailData struct {
	Subject string
	Title   string
	Body    string
}

func LoadConfig() (*Config, error) {
	configFile := os.Getenv("EMAIL_CONFIG_FILE")
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func SendEmail(config *Config, to, subject, body, attachmentPath string, isHTML bool) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	if isHTML {
		data := EmailData{
			Subject: subject,
			Title:   subject,
			Body:    body,
		}
		tmpl, err := template.ParseFiles("email-template.html")
		if err != nil {
			return err
		}
		htmlBody := ""
		htmlWriter := &strings.Builder{}
		err = tmpl.Execute(htmlWriter, data)
		if err != nil {
			return err
		}
		htmlBody = htmlWriter.String()
		m.SetBody("text/html", htmlBody)
	} else {
		m.SetBody("text/plain", body)
	}

	if attachmentPath != "" {
		m.Attach(attachmentPath)
	}

	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password)

	return d.DialAndSend(m)
}
