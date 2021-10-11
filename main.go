package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func check(wg *sync.WaitGroup, target string) {
	defer wg.Done()

	start := time.Now()
	_, err := http.Get(target)
	if err != nil {
		log.Println("Target :", target, "is not healthy ! ❌")
	} else {
		elapsed := time.Since(start)
		log.Println("Target :", target, "is healthy and responded in", elapsed, "! ✅")
	}
}

func main() {

	targets := []string{"http://localhost:8080", "https://google.fr", "https://padok.fr"}

	// Init a wait group
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go check(&wg, target)
	}

	// Wait for the waiting group to end
	wg.Wait()
}
