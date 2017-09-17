package main

import "time"
import "strconv"
import "fmt"
import "strings"
import "github.com/go-redis/redis"

var state_human_names map[string]string

type BadState struct {
	ServerName string
	Info       string
}

func GetBadStates(r *redis.Client, server_staleness_duration int) map[string]BadState {
	bad_states := map[string]BadState{}
	good_time := time.Now().Unix() - int64(server_staleness_duration)
	bad_servers, err := r.ZRevRangeByScoreWithScores(KeyMap["server_last_posts"], redis.ZRangeBy{Min: "0", Max: strconv.FormatInt(good_time, 10)}).Result()
	if err != nil {
		return bad_states
	}
	for _, item := range bad_servers {
		server := item.Member.(string)
		key := "stale_" + server
		last_reported_date := fmt.Sprintf("%s", time.Unix(int64(item.Score), 0))
		info := fmt.Sprintf("Server %s last reported on %v", server, last_reported_date)
		bad_states[key] = BadState{ServerName: server, Info: info}
	}
	return bad_states
}

func GetPrettyStateName(state_name string) string {
	parts := strings.Split(state_name, "_")
	if len(parts) == 0 {
		return ""
	}
	return state_human_names[parts[0]]
}

func init() {
	state_human_names = map[string]string{}
	state_human_names["stale"] = "Offline"
}
