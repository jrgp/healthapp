package main

import "fmt"
import "time"
import "encoding/json"
import "github.com/go-redis/redis"

func ServerLoadFromRedis(r *redis.Client, name string) (ServerExtendedInfo, error) {
	var info ServerExtendedInfo
	info_raw, err := r.Get(fmt.Sprintf(KeyMap["server_info"], name)).Result()
	if err != nil {
		return info, err
	}
	json.Unmarshal([]byte(info_raw), &info)
	info.Name = name

	score, err := r.ZScore(KeyMap["server_last_posts"], name).Result()
	if err == nil {
		info.LastUpdated = fmt.Sprintf("%s", time.Unix(int64(score), 0))
	}

	if info.Filesystems == nil {
		info.Filesystems = []ServerFilesystem{}
	}

	return info, nil
}
