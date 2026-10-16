package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "embed"

	apachelog "github.com/lestrrat-go/apache-logformat/v2"
)

//go:embed template.html
var homeTemplate string

var endpoints = map[string]string{
	"Downdetector Website": "https://downdetector.com",
	"Downdetector API":     "https://downdetectorapi.com/v2/ping",
}

type detectOrgHandler struct {
	tmpl        *template.Template
	endpointMap EndpointMap
}

func (doh detectOrgHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	doh.endpointMap.Lock.RLock()
	defer doh.endpointMap.Lock.RUnlock()
	err := doh.tmpl.Execute(w, doh.endpointMap)
	if err != nil {
		fmt.Printf("template: %v\n", err)
	}
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

	handler := apachelog.CombinedLog.Wrap(detectOrgHandler{
		tmpl:        tmpl,
		endpointMap: endpointMap,
	}, os.Stdout)

	http.Handle("/", handler)

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
