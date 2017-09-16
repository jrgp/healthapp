package main

var KeyMap map[string]string

func init() {
	KeyMap = map[string]string{}
	KeyMap["server_last_posts"] = "healthapp:server_last_posts"
	KeyMap["server_info"] = "healthapp:server_info:%s"
	KeyMap["server_alerts"] = "healthapp:server_alerts:%s"
	KeyMap["alert_currently_firing"] = "healthapp:alerts_current"
	KeyMap["alert_info"] = "healthapp:alert_info:%s"
	KeyMap["alerts_historical"] = "healthapp:alerts_list"
}
