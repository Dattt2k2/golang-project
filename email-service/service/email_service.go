package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/smtp"
	"os"

)

type EmailService struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("FROM_EMAIL"),
	}
}

func (s *EmailService) SendEmail(to, subject, templatePath string, data interface{}) error {

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var dataMap map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		dataMap = v
	case map[string]string:
		dataMap = make(map[string]interface{})
		for k, val := range v {
			dataMap[k] = val
		}
	default:
		b, _ := json.Marshal(data)
		_ = json.Unmarshal(b, &dataMap)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, dataMap); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body.String() + "\r\n")
	if err := smtp.SendMail(addr, auth, s.From, []string{to}, msg); err != nil {
		return err
	}
	return nil
}
