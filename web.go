package main

import "fmt"
import "path/filepath"
import "log"
import "net/http"
import "io/ioutil"
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

func GetServerStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func PostServerStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func ServersList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func AlertsList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func AlertInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func Web(r *redis.Client) {
	router := httprouter.New()

	// Backend api routes
	router.GET("/api/v0/status/:server_name", GetServerStatus)
	router.POST("/api/v0/status/:server_name", PostServerStatus)

	router.GET("/api/v0/servers", ServersList)
	router.GET("/api/v0/alerts", AlertsList)
	router.GET("/api/v0/alert/:alert_id", AlertInfo)

	// UI routes
	router.GET("/", SinglePageApp)
	router.GET("/static/*path", StaticResource)

	log.Fatal(http.ListenAndServe(":8080", router))
}
