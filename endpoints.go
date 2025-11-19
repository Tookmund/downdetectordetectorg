package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type EndpointStatus struct {
	Status    bool
	LastCheck *time.Time
	LastUp    *time.Time
}

type EndpointMap struct {
	TimestampFormat string
	UserAgent       string
	WaitTime        time.Duration
	Lock            *sync.RWMutex
	Map             map[string]EndpointStatus
}

// Should be run as a goroutine
func checkEndpoint(name string, url string, endpointMap EndpointMap) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("%s: Failed to create request: %v:", name, err)
	}
	req.Header.Add("User-Agent", endpointMap.UserAgent)

	for {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("%s: Failed to get %s: %v\n", name, url, err)
		}

		var endpointStatus EndpointStatus

		// Safe not to grab the read lock here, because only this thread will be modifying this map entry
		endpointStatus, ok := endpointMap.Map[name]
		if !ok {
			endpointStatus = EndpointStatus{}
		}

		if err != nil || resp.StatusCode != http.StatusOK {
			if endpointStatus.Status {
				endpointStatus.LastUp = endpointStatus.LastCheck
			}
			endpointStatus.Status = false
		} else {
			endpointStatus.Status = true
		}

		now := time.Now()
		endpointStatus.LastCheck = &now
		if endpointStatus.Status {
			endpointStatus.LastUp = &now
		}

		fmt.Printf("\n%s:\nStatus: %v\nLastCheck: %v\nLastUp: %v\n",
			name,
			endpointStatus.Status,
			endpointStatus.LastCheck,
			endpointStatus.LastUp,
		)

		endpointMap.Lock.Lock()
		endpointMap.Map[name] = endpointStatus
		endpointMap.Lock.Unlock()

		time.Sleep(endpointMap.WaitTime)
	}
}
