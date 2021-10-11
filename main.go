package main

import (
	"github.com/influxdata/influxdb-client-go/v2"
	"log"
	"net/http"
	"sync"
	"time"
)

func check(wg *sync.WaitGroup, target string, interval int) {

	// You can generate a Token from the "Tokens Tab" in the UI
	const token = "<token here>"
	const bucket = "go-monitoring-tool"
	const org = "fabienzucchet"

	client := influxdb2.NewClient("http://localhost:8086", token)
	// always close client at the end
	defer client.Close()

	// get non-blocking write client
	writeAPI := client.WriteAPI(org, bucket)

	defer wg.Done()

	for range time.Tick(time.Duration(interval) * time.Second) {
		start := time.Now()
		_, err := http.Get(target)
		if err != nil {
			log.Println("Target :", target, "is not healthy ! ❌")
			p := influxdb2.NewPointWithMeasurement("healthcheck").AddTag("target", target).AddField("status", 0).SetTime(time.Now())
			// write point asynchronously
			writeAPI.WritePoint(p)
		} else {
			elapsed := time.Since(start)
			log.Println("Target :", target, "is healthy and responded in", elapsed, "! ✅")
			p := influxdb2.NewPointWithMeasurement("healthcheck").AddTag("target", target).AddField("status", 1).AddField("response_time", float32(elapsed)).SetTime(time.Now())
			// write point asynchronously
			writeAPI.WritePoint(p)
		}
		// Flush writes
		writeAPI.Flush()
	}
}

func main() {

	targets := []string{"http://localhost:8080", "https://google.fr", "https://padok.fr"}

	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go check(&wg, target, 15)
	}

	wg.Wait()

}
