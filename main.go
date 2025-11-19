package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	_ "embed"
)

//go:embed template.html
var homeTemplate string

var endpoints = map[string]string{
	"Downdetector API": "https://downdetectorapi.com/v2/ping",
}

func main() {
	fmt.Println("Begin")

	tmpl, err := template.New("index").Parse(homeTemplate)
	if err != nil {
		fmt.Printf("failed to parse template: %v\n", err)
		return
	}
	endpointMap := EndpointMap{
		TimestampFormat: time.RFC3339,
		UserAgent:       "detectorg/1.0 detectorg@tookmund.com",
		WaitTime:        5 * time.Minute,
		Lock:            &sync.RWMutex{},
		Map:             make(map[string]EndpointStatus),
	}

	for name, url := range endpoints {
		go checkEndpoint(name, url, endpointMap)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		endpointMap.Lock.RLock()
		defer endpointMap.Lock.RUnlock()
		err := tmpl.Execute(w, endpointMap)
		if err != nil {
			fmt.Printf("template: %v\n", err)
		}
	})

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
