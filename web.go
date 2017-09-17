package main

import "fmt"
import "path/filepath"
import "log"
import "time"
import "net/http"
import "io/ioutil"
import "crypto/hmac"
import "crypto/sha512"
import "encoding/json"
import "encoding/base64"
import "github.com/julienschmidt/httprouter"
import "github.com/go-redis/redis"

var staticPath string
var mimes map[string]string

func init() {
	staticPath = "static"
	mimes = map[string]string{}
	mimes[".css"] = "text/css"
	mimes[".js"] = "text/javascript"
	mimes[".png"] = "image/png"
}

type App struct {
	r *redis.Client
}

func SinglePageApp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	path := staticPath + "/spa.html"
	contents, err := ioutil.ReadFile(path)
	if err == nil {
		fmt.Fprint(w, string(contents))
	} else {
		fmt.Fprint(w, "failed loading html page")
		log.Printf("Failed loading %v: %v", path, err)
	}
}

func StaticResource(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	path := staticPath + ps.ByName("path")
	ext := filepath.Ext(path)
	mime, _ := mimes[ext]
	contents, err := ioutil.ReadFile(path)
	if err == nil {
		if mime != "" {
			w.Header().Set("content-type", mime)
		}
		fmt.Fprint(w, string(contents))
	} else {
		fmt.Fprint(w, "failed loading file")
		log.Printf("Failed loading %v: %v", path, err)
	}
}

func (a *App) GetServerStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	server_name := ps.ByName("server_name")
	info, _ := ServerLoadFromRedis(a.r, server_name)
	w.Header().Set("content-type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(info)
}

func (a *App) PostServerStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	server_name := ps.ByName("server_name")
	hmac_header := r.Header.Get("X-INTEGRITY")
	hmac_raw, _ := base64.URLEncoding.DecodeString(hmac_header)
	body, _ := ioutil.ReadAll(r.Body)

	correct_hmac := hmac.New(sha512.New, []byte(Configs.ApiKey))
	correct_hmac.Write(body)

	if !hmac.Equal(hmac_raw, correct_hmac.Sum(nil)) {
		log.Printf("Received invalid hmac for server %s", server_name)
		http.Error(w, "invalid hmac", http.StatusUnauthorized)
		return
	}

	a.r.Set(fmt.Sprintf(KeyMap["server_info"], server_name), string(body), 0)
	a.r.ZAdd(KeyMap["server_last_posts"], redis.Z{Member: server_name, Score: float64(time.Now().Unix())})

	log.Printf("Received update for server %s", server_name)
}

func (a *App) ServersList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	serversResponseList := []ServerItemResponse{}

	servers, err := a.r.ZRangeWithScores(KeyMap["server_last_posts"], 0, -1).Result()
	if err != nil {
		return
	}

	good_time := float64(time.Now().Unix() - int64(Configs.ServerStalenessDuration))

	for _, item := range servers {
		server := ServerItemResponse{}
		server.Name = item.Member.(string)
		if item.Score > good_time {
			server.Good = true
		} else {
			server.Good = false
		}

		var more_info ServerExtendedInfo
		more_info, err = ServerLoadFromRedis(a.r, server.Name)
		if err == nil {
			server.Info = more_info
			server.Date = more_info.LastUpdated
			serversResponseList = append(serversResponseList, server)
		} else {
			log.Printf("Failed to load server %v: %v", server.Name, err)
		}
	}

	w.Header().Set("content-type", "application/json")
	serversResponse := ServerListResponse{Servers: serversResponseList}
	encoder := json.NewEncoder(w)
	encoder.Encode(serversResponse)
}

func (a *App) AlertsList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func (a *App) AlertInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func Web(r *redis.Client) {
	router := httprouter.New()

	app := App{r: r}

	// Backend api routes
	router.GET("/api/v0/status/:server_name", app.GetServerStatus)
	router.POST("/api/v0/status/:server_name", app.PostServerStatus)

	router.GET("/api/v0/servers", app.ServersList)
	router.GET("/api/v0/alerts", app.AlertsList)
	router.GET("/api/v0/alert/:alert_id", app.AlertInfo)

	// Static assets
	router.GET("/static/*path", StaticResource)

	// All of these output the same HTML file which acts according to the route
	// it got. Most of these are only hit on hard browser refreshes as html5
	// navigation api handles the rest
	router.GET("/", SinglePageApp)
	router.GET("/server/:server_name", SinglePageApp)
	router.GET("/alerts", SinglePageApp)
	router.GET("/alert/:alert_id", SinglePageApp)

	log.Fatal(http.ListenAndServe(Configs.ApiListen, router))
}
