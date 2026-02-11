package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

type Conf struct {
	SenderEmail string `yaml:"senderEmail"`
	Password    string `yaml:"password"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
}

type Message struct {
	From        string
	To          string
	Subject     string
	Description string
}

const emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}

{{.Description}}
`

var tempEmail = template.Must(template.New("email").Parse(emailTemplate))

var Send = func(conf *Conf, msg *Message, body []byte) error {
	auth := smtp.PlainAuth("", conf.SenderEmail, conf.Password, conf.Host)
	confServer := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	if err := smtp.SendMail(confServer, auth, msg.From, []string{msg.To}, body); err != nil {
		return fmt.Errorf("SendMail: %w", err)
	}

	return nil
}

func Alert(conf *Conf, msg *Message) error {
	var body bytes.Buffer

	msg.From = conf.SenderEmail
	if err := tempEmail.Execute(&body, msg); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := Send(conf, msg, body.Bytes()); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}
