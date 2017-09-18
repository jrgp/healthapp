package main

import "time"
import "strconv"
import "errors"
import "fmt"
import "log"
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
	if err == nil {
		alert.ID = StateName + "_" + generated_uuid.String()
		alert.SaveNewAlert(r)
		if Configs.EnableEmails {
			NotifyAlertNew(alert)
		}
	} else {
		log.Printf("Failed generating uuid for new alert %v: %v", StateName, err)
	}
}

func (alert Alert) Close(r *redis.Client) {
	alert.EndTime = time.Now().Unix()
	alert.Duration = alert.EndTime - alert.StartTime
	alert.SaveClosedAlert(r)
	if Configs.EnableEmails {
		NotifyAlertClosed(alert)
	}
}

func (alert Alert) SaveNewAlert(r *redis.Client) {
	values := make(map[string]interface{})
	values["start_time"] = strconv.FormatInt(alert.StartTime, 10)
	values["end_time"] = strconv.FormatInt(alert.EndTime, 10)
	values["info"] = alert.Description
	values["server_name"] = alert.ServerName
	values["state_name"] = alert.StateName

	r.HMSet(fmt.Sprintf(KeyMap["alert_info"], alert.ID), values)
	r.ZAdd(KeyMap["alerts_historical"], redis.Z{Score: float64(alert.StartTime), Member: alert.ID})
	r.ZAdd(fmt.Sprintf(KeyMap["server_alerts"], alert.ServerName), redis.Z{Score: float64(alert.StartTime), Member: alert.ID})
	r.HSet(KeyMap["alert_currently_firing"], alert.StateName, alert.ID)
}

func (alert Alert) SaveClosedAlert(r *redis.Client) {
	r.HDel(KeyMap["alert_currently_firing"], alert.StateName)
	alert_key := fmt.Sprintf(KeyMap["alert_info"], alert.ID)
	exists, _ := r.Exists(alert_key).Result()

	if exists == 0 {
		return
	}

	values := make(map[string]interface{})
	values["end_time"] = strconv.FormatInt(alert.EndTime, 10)
	values["start_time"] = strconv.FormatInt(alert.StartTime, 10)
	values["duration"] = strconv.FormatInt(alert.Duration, 10)
	r.HMSet(fmt.Sprintf(KeyMap["alert_info"], alert.ID), values)
}

func (alert Alert) GetPrettyRepresentation(r *redis.Client) PrettyAlertInfo {
	info := PrettyAlertInfo{}
	info.StartTime = fmt.Sprintf("%s", time.Unix(int64(alert.StartTime), 0))
	if alert.EndTime == 0 {
		info.EndTime = "Ongoing"
		duration := time.Duration(time.Now().Unix()-alert.StartTime) * time.Second
		info.Duration = duration.String()
	} else {
		info.EndTime = fmt.Sprintf("%s", time.Unix(int64(alert.EndTime), 0))
		duration := time.Duration(alert.EndTime-alert.StartTime) * time.Second
		info.Duration = duration.String()
	}
	info.ID = alert.ID
	info.StateName = alert.StateName
	info.Description = alert.Description
	info.HumanBadName = GetPrettyStateName(alert.StateName)
	if alert.EndTime > 0 {
		info.Ongoing = false
	} else {
		info.Ongoing = true
	}
	server, err := ServerLoadFromRedis(r, alert.ServerName)
	if err == nil {
		info.Server = server
	} else {
		log.Printf("failed looking up server %v", alert.ServerName)
	}
	return info
}

func LoadAlertFromRedis(r *redis.Client, alert_id string) (Alert, error) {
	alert_key := fmt.Sprintf(KeyMap["alert_info"], alert_id)
	alert_info, err := r.HGetAll(alert_key).Result()
	if err != nil {
		return Alert{ID: alert_id}, err
	}
	if len(alert_info) == 0 {
		return Alert{ID: alert_id}, errors.New("alert not found")
	}
	start_time, _ := strconv.ParseInt(alert_info["start_time"], 10, 64)
	end_time, _ := strconv.ParseInt(alert_info["end_time"], 10, 64)
	return Alert{
		StateName:   alert_info["state_name"],
		ID:          alert_id,
		EndTime:     end_time,
		StartTime:   start_time,
		ServerName:  alert_info["server_name"],
		Description: alert_info["info"],
	}, nil
}
