package main

import "time"
import "strconv"
import "fmt"
import "github.com/go-redis/redis"
import "github.com/nu7hatch/gouuid"

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
