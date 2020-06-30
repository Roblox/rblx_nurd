package main

import (
	"log"
	"time"
)

// Configure the cluster addresses and frequency of monitor
func Config() ([]string, int, time.Duration) {
	addresses := []string{"***REMOVED***:***REMOVED***"} // one address per cluster
	buffer := len(addresses)
	duration, errParseDuration := time.ParseDuration("1m")

	if errParseDuration != nil {
		log.Fatal("Error:", errParseDuration)
	}
	
	return addresses, buffer, duration
}