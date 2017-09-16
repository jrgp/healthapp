package main

import "fmt"
import "path/filepath"
import "log"
import "net/http"
import "io/ioutil"
import "encoding/json"
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
}

func (a *App) PostServerStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func (a *App) ServersList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	serversResponseList := []ServerItemResponse{}

	servers, err := a.r.ZRangeWithScores(KeyMap["server_last_posts"], 0, -1).Result()
	if err != nil {
		return
	}

	for _, item := range servers {
		server := ServerItemResponse{}
		server.Name = item.Member.(string)
		if item.Score > float64(Configs.ServerStalenessDuration) {
			server.Good = true
		} else {
			server.Good = false
		}

		var more_info ServerExtendedInfo
		var more_info_raw string
		more_info_raw, err = a.r.Get(fmt.Sprintf(KeyMap["server_info"], server.Name)).Result()
		json.Unmarshal([]byte(more_info_raw), &more_info)
		more_info.Name = server.Name
		server.Info = more_info

		serversResponseList = append(serversResponseList, server)
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

	log.Fatal(http.ListenAndServe(":8080", router))
}
