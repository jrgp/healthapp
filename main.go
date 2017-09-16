package main

import "fmt"
import "os"
import "flag"
import "github.com/go-redis/redis"

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
