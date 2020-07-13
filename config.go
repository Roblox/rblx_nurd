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

func loadConfig(path string) ([]string, string, int, time.Duration) {
	var metricsAddress string
	var nomadAddresses []string
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

	metricsAddress = config.VictoriaMetrics.URL + ":" + config.VictoriaMetrics.Port
	nomadAddresses = make([]string, 0)

	for _, server := range config.Nomad {
		nomadAddresses = append(nomadAddresses, server.URL+":"+server.Port)
	}

	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Fatal(err)
	}

	return nomadAddresses, metricsAddress, len(nomadAddresses), duration
}
