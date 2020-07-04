package main

import (
	"bufio"
	"log"
	"os"
	"time"
)

func Config(path string) ([]string, string, int, time.Duration) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	metricsAddress := scanner.Text()

	addresses := make([]string, 0)
    for scanner.Scan() {
        addresses = append(addresses, scanner.Text())
    }

	buffer := len(addresses)
	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Fatal(err)
	}

	return addresses, metricsAddress, buffer, duration
}
