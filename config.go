package main

import "log"
import "os"
import "fmt"
import "io/ioutil"
import "gopkg.in/yaml.v2"

type Config struct {
	Redis                   string   `yaml:"redis"`
	ApiKey                  string   `yaml:"api_key"`
	ApiURL                  string   `yaml:"api_url"`
	SetUID                  string   `yaml:"setuid"`
	SetGID                  string   `yaml:"setgid"`
	ServerStalenessDuration int      `yaml:"server_staleness_duration"`
	AlertProcessInterval    int      `yaml:"alert_process_interval"`
	AlertSendEmailInterval  int      `yaml:"alert_send_email_interval"`
	EnableEmails            bool     `yaml:"enable_emails"`
	EmailServer             string   `yaml:"email_server"`
	EmailSender             string   `yaml:"email_sender"`
	EmailRecipients         []string `yaml:"email_recipients"`
}

var Configs Config

func init() {
	paths := []string{
		os.Getenv("CONFIG_FILE"),
		"/etc/healthapp/config.yaml",
		"/usr/local/healthapp/config.yaml",
		"config.yaml",
	}

	var contents []byte
	contents = nil

	for _, path := range paths {
		if path == "" {
			continue
		}

		this_contents, err := ioutil.ReadFile(path)

		if err == nil {
			contents = this_contents
			fmt.Println("Reading config from", path)
			break
		}
	}

	if contents == nil {
		fmt.Println("Could not find any config files to load. Try setting CONFIG_FILE to its path")
		os.Exit(1)
	}

	err := yaml.Unmarshal(contents, &Configs)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
