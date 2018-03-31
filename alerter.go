package main

import "github.com/go-redis/redis"
import "log"
import "time"

func Alerter(r *redis.Client) {
	sleep_time := 60
	for {
		err := alert_run(r)
		if err == nil {
			log.Printf("Alert processor run done. Will sleep %v seconds", sleep_time)
		} else {
			log.Printf("Alert run problem: %v. Will sleep %v seconds", err, sleep_time)
		}
		time.Sleep(time.Duration(sleep_time) * time.Second)
	}
}

func alert_run(r *redis.Client) error {
	closed_alerts := 0
	ongoing_alerts := 0
	new_alerts := 0

	bad_states := map[string]BadState{}
	for key, state := range GetStaleHosts(r, Configs.ServerStalenessDuration) {
		bad_states[key] = state
	}

	for key, state := range GetBadDiskStates(r, Configs.ServerStalenessDuration, Configs.MaxFilesystemPercentage) {
		bad_states[key] = state
	}

	currently_firing, err := r.HGetAll(REDIS_KEY_ALERT_CURRENTLY_FIRING).Result()
	if err != nil {
		return err
	}

	// 1: iterate through mapping of currently firing alerts in redis, checking if each
	// is stil in bad state. if not mark them as closed
	for state_name, alert_id := range currently_firing {
		if _, ok := bad_states[state_name]; ok {
			log.Printf("Alert %s still firing: %s", state_name, alert_id)
			// TODO: ongoing notification
			ongoing_alerts++
		} else {
			log.Printf("Alert %s no longer firing. Closing.", state_name)
			alert, _ := LoadAlertFromRedis(r, alert_id)
			alert.Close(r)
			closed_alerts++
		}

		// purge from bad_states
		delete(bad_states, state_name)
	}

	// 2: create new alerts for states which are bad but not yet kept track of
	for state_name, description := range bad_states {
		alert := Alert{}
		alert.Create(r, state_name, description.Info, description.ServerName)
		new_alerts++
	}

	// 3 (TODO): purge records of ancient alerts
	log.Printf("New alerts: %v. Ongoing alerts: %v. Closed alerts: %v", new_alerts, ongoing_alerts, closed_alerts)

	return nil
}

// vim: filetype=go
