package main

import "fmt"
import "os"
import "log"
import "time"
import "strconv"
import "flag"
import "github.com/nu7hatch/gouuid"
import "github.com/go-redis/redis"

var KeyMap map[string]string


type BadState struct {
	ServerName string
	Info       string
}

func GetBadStates(r *redis.Client, server_staleness_duration int64) map[string]BadState {
	bad_states := map[string]BadState{}
	good_time := time.Now().Unix() - server_staleness_duration
	bad_servers, err := r.ZRevRangeByScoreWithScores(KeyMap["server_last_posts"], redis.ZRangeBy{Max: "0", Min: strconv.FormatInt(good_time, 10)}).Result()
	if err == nil {
		return bad_states
	}
	for _, item := range bad_servers {
		server := item.Member.(string)
		key := "stale_" + server
		info := fmt.Sprintf("Server %s last reported on %v", server, item.Score)
		bad_states[key] = BadState{ServerName: server, Info: info}
	}
	return bad_states
}

type Alert struct {
	StartTime   int64
	EndTime     int64
	Duration    int64
	ID          string
	StateName   string
	Description string
	ServerName  string
}

func (alert Alert) Create(r *redis.Client, StateName, Description, ServerName string) {
	alert.StartTime = time.Now().Unix()
	alert.EndTime = 0
	alert.StateName = StateName
	alert.Description = Description
	alert.ServerName = ServerName

	generated_uuid, err := uuid.NewV4()
	if err != nil {
		alert.ID = StateName + "_" + generated_uuid.String()
		alert.SaveNewAlert(r)
	}
}

func (alert Alert) Close(r *redis.Client) {
	alert.EndTime = time.Now().Unix()
	alert.Duration = alert.EndTime - alert.StartTime
	alert.SaveClosedAlert(r)
	// TODO notification
}

func (alert Alert) SaveNewAlert(r *redis.Client) {
	values := make(map[string]interface{})
	values["start_time"] = strconv.FormatInt(alert.StartTime, 10)
	values["end_time"] = strconv.FormatInt(alert.EndTime, 10)
	values["info"] = alert.Description
	values["server_name"] = alert.ServerName

	r.HMSet(fmt.Sprintf(KeyMap["alert_info"], alert.ID), values)
	r.ZAdd(KeyMap["alerts_historical"], redis.Z{Score: float64(alert.StartTime), Member: alert.ID})
	r.ZAdd(fmt.Sprintf(KeyMap["server_alerts"], alert.ServerName), redis.Z{Score: float64(alert.StartTime), Member: alert.ID})
	r.HSet(KeyMap["alert_currently_firing"], alert.StateName, alert.ID)
}

func (alert Alert) SaveClosedAlert(r *redis.Client) {
	r.HDel(KeyMap["alert_currently_firing"], alert.StateName)
	alert_key := fmt.Sprintf(KeyMap["alert_info"], alert.ID)
	exists, _ := r.Exists(alert_key).Result()

	if exists == 1 {
		return
	}

	values := make(map[string]interface{})
	values["end_time"] = strconv.FormatInt(alert.EndTime, 10)
	values["start_time"] = strconv.FormatInt(alert.StartTime, 10)
	values["duration"] = strconv.FormatInt(alert.Duration, 10)
	r.HMSet(fmt.Sprintf(KeyMap["alert_info"], alert.ID), values)

}

func LoadAlertFromRedis(r *redis.Client, state_name, alert_id string) Alert {
	alert_key := fmt.Sprintf(KeyMap["alert_info"], alert_id)
	alert_info, _ := r.HGetAll(alert_key).Result()
	start_time, _ := strconv.ParseInt(alert_info["end_time"], 10, 64)
	end_time, _ := strconv.ParseInt(alert_info["start_time"], 10, 64)
	return Alert{
		StateName: state_name,
		ID:        alert_id,
		EndTime:   end_time,
		StartTime: start_time,
	}
}

func init() {
	KeyMap = map[string]string{}
	KeyMap["server_last_posts"] = "healthapp:server_last_posts"
	KeyMap["server_info"] = "healthapp:server_info:%s"
	KeyMap["server_alerts"] = "healthapp:server_alerts:%s"
	KeyMap["alert_currently_firing"] = "healthapp:alerts_current"
	KeyMap["alert_info"] = "healthapp:alert_info:%s"
	KeyMap["alerts_historical"] = "healthapp:alerts_list"
}

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

func Agent(r *redis.Client) {

}

func Web(r *redis.Client) {

}

func alert_run(r *redis.Client) error {
	closed_alerts := 0
	ongoing_alerts := 0
	new_alerts := 0

	bad_states := GetBadStates(r, 300)

	currently_firing, err := r.HGetAll(KeyMap["alert_currently_firing"]).Result()
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
			alert := LoadAlertFromRedis(r, state_name, alert_id)
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

func get_redis() *redis.Client {
  return redis.NewClient(&redis.Options{
    Addr:     "127.0.0.1:6379",
    Password: "", // no password set
    DB:       0,  // use default DB
  })
}

func main() {

  var run_alerter bool
  var run_web bool
  var run_agent bool

  flag.BoolVar(&run_alerter, "alerter", false, "run alerter")
  flag.BoolVar(&run_web, "serve", false, "serve web api")
  flag.BoolVar(&run_agent, "agent", false, "run agent")

  flag.Parse()

  client := get_redis()

  if run_alerter {
    Alerter(client)
  } else if run_web {
    Web(client)
  } else if run_agent {
    Agent(client)
  } else {
    fmt.Println("No sub command to run. See -h for usage.")
    os.Exit(1)
  }
}

// vim: filetype=go
