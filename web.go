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

func Web(r *redis.Client) {
	staticPath = "static"
	mimes = map[string]string{}
	mimes[".css"] = "text/css"
	mimes[".js"] = "text/javascript"

	router := httprouter.New()
	router.GET("/", SinglePageApp)
	router.GET("/static/*path", StaticResource)
	log.Fatal(http.ListenAndServe(":8080", router))
}
