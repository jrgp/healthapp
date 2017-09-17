package main

var message_templates map[string]MessageTemplate

type MessageTemplate struct {
	subject string
	body    string
}

func NotifyAlertNew() {

}

func NotifyAlertClosed() {

}

func NotifyAlertOngoing() {

}

func SendEmail(subject, body string) error {
	return nil
}

func init() {
	message_templates = map[string]MessageTemplate{}

	message_templates["new_alert"] = MessageTemplate{
		subject: "New Alert \"%s\"!",
		body:    "Hi,\n\nAlert \"%s\" has just started firing.\n\n%s\n\nRegards",
	}

	message_templates["closed_alert"] = MessageTemplate{
		subject: "",
		body:    "",
	}

	message_templates["ongoing_alert"] = MessageTemplate{
		subject: "",
		body:    "",
	}
}
