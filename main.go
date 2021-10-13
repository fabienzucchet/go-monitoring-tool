package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
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
			p := influxdb2.NewPointWithMeasurement("healthcheck").AddTag("target", target).AddField("status", 1).AddField("response_time", float32(elapsed)).AddField("status_code", resp.StatusCode).SetTime(time.Now())
			// write point asynchronously
			writeAPI.WritePoint(p)
		}
		// Flush writes
		writeAPI.Flush()
	}
}

func query(influxConfig *InfluxConfig) {
	query := fmt.Sprintf("from(bucket:\"%v\")|> range(start: -1h) |> filter(fn: (r) => r._measurement == \"healthcheck\")", influxConfig.bucket)

	client := influxdb2.NewClient(influxConfig.url, influxConfig.token)
	// always close client at the end
	defer client.Close()

	// Get query client
	queryAPI := client.QueryAPI(influxConfig.org)

	// get QueryTableResult
	result, err := queryAPI.Query(context.Background(), query)

	if err == nil {
	// Iterate over query response
	for result.Next() {
		// Notice when group key has changed
		if result.TableChanged() {
		fmt.Printf("table: %s\n", result.TableMetadata().String())
		}
		// Access data
		fmt.Println("time:", result.Record().Time(), "value:", result.Record().Value())
	}
	// check for an error
	if result.Err() != nil {
		fmt.Printf("query parsing error: %\n", result.Err().Error())
	}
	} else {
		panic(err)
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

	log.Println("Started the monitoring tool. 🏁")

	// HTTP handlers
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query(&influxConfig)
        tmpl.Execute(w, "lol")
    })

	// Serve static files
	fs := http.FileServer(http.Dir("static/"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start HTTP server
	http.ListenAndServe(":8081", nil)

}
