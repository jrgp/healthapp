package main

import "strings"
import "time"
import "os"
import "log"
import "fmt"
import "net/http"
import "encoding/json"
import "encoding/base64"
import "crypto/hmac"
import "crypto/sha512"
import "github.com/matishsiao/goInfo"

func generatePayload(info ServerExtendedInfo, api_key string) (string, string) {
	body, _ := json.Marshal(info)
	mac := hmac.New(sha512.New, []byte(api_key))
	mac.Write(body)
	sum := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return string(body), sum
}

func postPayload(url string, body string, hmac_header string) error {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Add("X-INTEGRITY", hmac_header)

	if err != nil {
		return fmt.Errorf("Failed creating http client: %v", err)
	}

	var resp *http.Response
	resp, err = client.Do(req)

	if err != nil {
		return fmt.Errorf("Failed performing post request: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Got bad status code from api: %v", resp.StatusCode)
	}

	return nil

}

func Agent() {

	hostname, _ := os.Hostname()
	gi := goInfo.GetInfo()
	myInfo := ServerExtendedInfo{
		OS:     strings.Title(gi.Kernel),
		Kernel: gi.Core,
		Name:   hostname,
	}

	api_url := Configs.ApiURL + "/api/v0/status/" + hostname
	sleep_time := 30

	for {
		body, macstr := generatePayload(myInfo, Configs.ApiKey)
		err := postPayload(api_url, body, macstr)
		if err == nil {
			log.Printf("Successful post. Sleeping for %v seconds.", sleep_time)
		} else {
			log.Printf("%v. Sleeping for %v seconds.", err, sleep_time)
		}
		time.Sleep(time.Duration(sleep_time) * time.Second)
	}
}
