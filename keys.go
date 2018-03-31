package main

const (
	REDIS_KEY_SERVER_LASTS_POSTS     = "healthapp:server_last_posts"
	REDIS_KEY_SERVER_INFO            = "healthapp:server_info:%s"
	REDIS_KEY_SERVER_ALERTS          = "healthapp:server_alerts:%s"
	REDIS_KEY_ALERT_CURRENTLY_FIRING = "healthapp:alerts_current"
	REDIS_KEY_ALERT_INFO             = "healthapp:alert_info:%s"
	REDIS_KEY_ALERT_HISTORICAL       = "healthapp:alerts_list"
)
