package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxConfig struct {
	token string
	bucket string
	org string
	url string
}


func check(target string, interval int, influxConfig *InfluxConfig) {

	client := influxdb2.NewClient(influxConfig.url, influxConfig.token)
	// always close client at the end
	defer client.Close()

	// get non-blocking write client
	writeAPI := client.WriteAPI(influxConfig.org, influxConfig.bucket)

	for range time.Tick(time.Duration(interval) * time.Second) {
		start := time.Now()
		resp, err := http.Get(target)
		if err != nil {
			log.Println("Target :", target, "is not healthy ! ❌")
			p := influxdb2.NewPointWithMeasurement("healthcheck").AddTag("target", target).AddField("status", 0).SetTime(time.Now())
			// write point asynchronously
			writeAPI.WritePoint(p)
		} else {
			elapsed := time.Since(start)
			log.Println("Target :", target, "is healthy and responded in", elapsed, "! ✅")
			p := influxdb2.NewPointWithMeasurement("healthcheck").AddTag("target", target).AddField("status", 1).AddField("response_time", float32(elapsed)).AddField("status_code", int(resp.StatusCode)).SetTime(time.Now())
			// write point asynchronously
			writeAPI.WritePoint(p)
		}
		// Flush writes
		writeAPI.Flush()
	}
}

func main() {

	influxConfig := InfluxConfig{
		token: os.Getenv("TSDB_TOKEN"),
		bucket: "go-monitoring-tool",
		org: "fabienzucchet",
		url: "http://localhost:8086",
	}

	targets := []string{"http://localhost:8080", "https://google.fr", "https://padok.fr"}

	log.Println("Starting the monitoring tool... ⌛️")

	for _, target := range targets {
		go check(target, 15, &influxConfig)
	}

	log.Println("Started the monitoring tool. ⚡️")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()

}
