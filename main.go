package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	bson "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"
)

type InfluxConfig struct {
	token string
	bucket string
	org string
	url string
}

type Target struct {
	Url string `json:"url"`
	CollectionInterval int `json:"collectioninterval,string"`
}

type CustomResponseBody struct {
	Status string `json:"status"`
	Message string `json:"message"`
}

type Point struct{
	X int64 `json:"x"`
	Y string `json:"y"`
}

type AvailabilityData struct {
	Target string `json:"target"`
	Availability string `json:"availability"`	
}

type StatusCodeData map[string]int

func check(target string, interval int, influxConfig *InfluxConfig) {

	log.Println("Checking target", target, "with collection interval", interval, "s 🔍")

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
			elapsed := time.Since(start).Milliseconds()
			if resp.StatusCode >= 400 {
				log.Println("Target :", target, "is unhealthy and responded in", elapsed, "ms ! ❌")
			} else {
				log.Println("Target :", target, "is healthy and responded in", elapsed, "ms ! ✅")
			}
			p := influxdb2.NewPointWithMeasurement("healthcheck").AddTag("target", target).AddField("status", 1).AddField("response_time", float32(elapsed)).AddField("status_code", resp.StatusCode).SetTime(time.Now())
			// write point asynchronously
			writeAPI.WritePoint(p)
		}
		// Flush writes
		writeAPI.Flush()
	}
}

func executeFluxQuery(influxConfig *InfluxConfig, query string) (*api.QueryTableResult, error) {

	client := influxdb2.NewClient(influxConfig.url, influxConfig.token)
	// always close client at the end
	defer client.Close()

	// Get query client
	queryAPI := client.QueryAPI(influxConfig.org)

	// get QueryTableResult
	return queryAPI.Query(context.Background(), query)
}

func insertTarget(collection *mongo.Collection, target *Target) error {
	_, err := collection.InsertOne(context.TODO(), target)
	
	return err
}

func readWebsites(collection *mongo.Collection) []Target {
	findOptions := options.Find()
	
	var results []Target

	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(context.TODO()) {
    
		// create a value into which the single document can be decoded
		var elem Target
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
	
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	
	// Close the cursor once finished
	cur.Close(context.TODO())

	return results
}

func main() {

	log.Println("Starting the monitoring tool... ⌛️")

	// INFLUXDB //
	influxConfig := InfluxConfig{
		token: os.Getenv("TSDB_TOKEN"),
		bucket: "go-monitoring-tool",
		org: "fabienzucchet",
		url: "http://localhost:8086",
	}

	// MONGODB //
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://user:password@localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Connected to MongoDB! 💾")

	collection := client.Database("healthcheck").Collection("targets")

	// Fetch the websites already monitored
	targets := readWebsites(collection)

	for _, target := range targets {
		go check(target.Url, target.CollectionInterval, &influxConfig)
	}

	log.Println("Started the monitoring tool. 🏁")

	// HTTP HANDLERS //

	// /target/
	// POST : Register a new target in the database
	http.HandleFunc("/target", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			decoder := json.NewDecoder(r.Body)
			var target Target

			err := decoder.Decode(&target)
			if err != nil {
				json.NewEncoder(w).Encode(CustomResponseBody{Status: "error", Message: "An error occured !"})
				w.WriteHeader(http.StatusInternalServerError)
			}
			
			err = insertTarget(collection, &target)

			if err != nil {
				json.NewEncoder(w).Encode(CustomResponseBody{Status: "error", Message: "An error occured !"})
				w.WriteHeader(http.StatusInternalServerError)
			}

			go check(target.Url, target.CollectionInterval, &influxConfig)

			json.NewEncoder(w).Encode(CustomResponseBody{Status: "success", Message: "Target successfully added !"})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /metrics/latency
	http.HandleFunc("/metrics/latency", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
		    duration, ok := r.URL.Query()["duration"]
			if !ok || len(duration[0]) < 1 {
				duration = []string{"-1h"}
			}

			// Query data in TSDB
			query := fmt.Sprintf("from(bucket:\"%s\")|> range(start: %s) |> filter(fn: (r) => r._measurement == \"healthcheck\" and r._field == \"response_time\")", influxConfig.bucket, duration[0])
			result, err := executeFluxQuery(&influxConfig, query)

			if err != nil {
				http.Error(w, "Error fetching data in TSDB", http.StatusBadRequest)
			}

			dataMap := map[string][]Point{}

			for result.Next() {
				target := fmt.Sprintf("%v", result.Record().ValueByKey("target"))
				dataMap[target] = append(dataMap[target], Point{X: result.Record().Time().Unix() ,Y: fmt.Sprintf("%v", result.Record().Value())})
			}

			json.NewEncoder(w).Encode(dataMap)

		default:
		    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /metrics/availability
	http.HandleFunc("/metrics/availability", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
		    duration, ok := r.URL.Query()["duration"]
			if !ok || len(duration[0]) < 1 {
				duration = []string{"-1h"}
			}

			// Query data in TSDB
			query := fmt.Sprintf("from(bucket:\"%v\")|> range(start: %s) |> filter(fn: (r) => r._measurement == \"healthcheck\" and r._field == \"status\") |> mean()", influxConfig.bucket, duration[0])
			result, err := executeFluxQuery(&influxConfig, query)

			if err != nil {
				http.Error(w, "Error fetching data in TSDB", http.StatusBadRequest)
			}

			var data []AvailabilityData

			for result.Next() {
				data = append(data, AvailabilityData{
					Target: fmt.Sprintf("%v", result.Record().ValueByKey("target")),
					Availability: fmt.Sprintf("%v", result.Record().Value()),
				})
			}

			json.NewEncoder(w).Encode(data)

		default:
		    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /metrics/httpstatus
	http.HandleFunc("/metrics/httpstatus", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
		    duration, ok := r.URL.Query()["duration"]
			if !ok || len(duration[0]) < 1 {
				duration = []string{"-1h"}
			}

			// Query data in TSDB
			query := fmt.Sprintf("from(bucket:\"%v\")|> range(start: %s) |> filter(fn: (r) => r._measurement == \"healthcheck\" and r._field == \"status_code\")", influxConfig.bucket, duration[0])
			result, err := executeFluxQuery(&influxConfig, query)

			if err != nil {
				http.Error(w, "Error fetching data in TSDB", http.StatusBadRequest)
			}

			dataMap := map[string]StatusCodeData{}

			for result.Next() {
				target := fmt.Sprintf("%v", result.Record().ValueByKey("target"))
				value := fmt.Sprintf("%v", result.Record().Value())

				if _,ok := dataMap[target]; !ok {
					dataMap[target] = StatusCodeData{}
				}

				if _, ok := dataMap[target][value]; ok {
					dataMap[target][value] = dataMap[target][value] + 1
				} else {
					dataMap[target][value] = 1
				}
			}

			json.NewEncoder(w).Encode(dataMap)

		default:
		    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// TEMPLATES //
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl.Execute(w, "lol")
    })

	// STATIC FILES //
	fs := http.FileServer(http.Dir("static/"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start HTTP server
	http.ListenAndServe(":8081", nil)

}

// switch r.Method {
// case http.MethodGet:
//     // Serve the resource.
// case http.MethodPost:
//     // Create a new record.
// case http.MethodPut:
//     // Update an existing record.
// case http.MethodDelete:
//     // Remove the record.
// default:
//     http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// }
