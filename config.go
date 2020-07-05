package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type ConfigFile struct {
	VictoriaMetrics Server
	Nomad           []Server
}

type Server struct {
	URL  string
	Port string
}

func Config(path string) ([]string, string, int, time.Duration) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	byte, err := ioutil.ReadAll(file)
	var config ConfigFile
	err = json.Unmarshal(byte, &config)
	if err != nil {
		log.Fatal(err)
	}

	metricsAddress := config.VictoriaMetrics.URL + ":" + config.VictoriaMetrics.Port
	nomadAddresses := make([]string, 0)

	nomad := config.Nomad
	for _, server := range nomad {
		nomadAddresses = append(nomadAddresses, server.URL+":"+server.Port)
	}

	buffer := len(nomadAddresses)
	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Fatal(err)
	}

	return nomadAddresses, metricsAddress, buffer, duration
}
