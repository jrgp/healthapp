package main

import "fmt"
import "log"
import "time"
import "gopkg.in/gomail.v2"

var message_templates map[string]MessageTemplate

type MessageTemplate struct {
	Subject string
	Body    string
}

func NotifyAlertNew(alert Alert) {
	subject := fmt.Sprintf(message_templates["new_alert"].Subject, alert.StateName)
	body := fmt.Sprintf(message_templates["new_alert"].Body, alert.StateName, alert.ID)
	go SendEmail(subject, body)
}

func NotifyAlertClosed(alert Alert) {
	subject := fmt.Sprintf(message_templates["closed_alert"].Subject, alert.StateName)
	duration := time.Duration(alert.Duration) * time.Second
	body := fmt.Sprintf(message_templates["closed_alert"].Body, alert.StateName, duration.String(), alert.ID)
	go SendEmail(subject, body)
}

func NotifyAlertOngoing(alert Alert) {
	subject := fmt.Sprintf(message_templates["ongoing_alert"].Subject, alert.StateName)
	body := fmt.Sprintf(message_templates["ongoing_alert"].Body, alert.StateName, alert.ID)
	go SendEmail(subject, body)
}

func SendEmail(subject, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", Configs.EmailSender)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	for _, address := range Configs.EmailRecipients {
		m.SetHeader("To", address)
	}

	d := gomail.Dialer{Host: Configs.EmailServer, Port: Configs.EmailServerPort}
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed emailing: %s", err)
	} else {
		log.Printf("Emailed '%s' notification", subject)
	}
}

func init() {
	message_templates = map[string]MessageTemplate{}

	message_templates["new_alert"] = MessageTemplate{
		Subject: "New Alert \"%s\"!",
		Body:    "Hi,\n\nAlert \"%s\" has just started firing.\n\n%s\n\nRegards",
	}

	message_templates["ongoing_alert"] = MessageTemplate{
		Subject: "Alert \"%s\" still firing",
		Body:    "Hi,\n\nAlert \"%s\" is still firing.\n\n%s\n\nRegards",
	}

	message_templates["closed_alert"] = MessageTemplate{
		Subject: "Alert \"%s\" closed",
		Body:    "Hi,\n\nAlert \"%s\" closed after %s.\n\n%s\n\nRegards",
	}
}
