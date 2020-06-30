package main

import (
	"log"
	"time"
)

// Configure the cluster addresses and frequency of monitor
func Config() ([]string, int, time.Duration) {
	addresses := []string{"***REMOVED***:***REMOVED***"} // one address per cluster
	buffer := len(addresses)
	duration, err := time.ParseDuration("1m")

	if err != nil {
		log.Fatal("Error:", err)
	}

	return addresses, buffer, duration
}
