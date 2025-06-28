package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

type EmailService struct {
	Host string
	Port string 
	Username string
	Password string
	From string
}

func NewEmailService() *EmailService {
	return &EmailService{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From: os.Getenv("FROM_EMAIL"),
	}
}


func (s *EmailService) SendEmail (to, subject, templatePath string, data interface{}) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err 
	}

	var body bytes.Buffer 
	if err := tmpl.Execute(&body, data); err != nil {
		return err 
	}

	addr := fmt.Sprint("%s:%s", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body.String() + "\r\n")
	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}