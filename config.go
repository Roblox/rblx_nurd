package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

type ConfigFile struct {
	VictoriaMetrics Server
	Nomad           []Server
}

type Server struct {
	URL  string
	Port string
}

var (
	nomadAddresses []string
	metricsAddress string
)

func loadConfig(path string) error {
	nomadAddresses = []string{}

	logFile, err := os.OpenFile("nurd.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error(err)
	}
	log.SetOutput(logFile)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var config ConfigFile
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	metricsAddress = config.VictoriaMetrics.URL + ":" + config.VictoriaMetrics.Port

	for _, server := range config.Nomad {
		nomadAddresses = append(nomadAddresses, server.URL+":"+server.Port)
	}

	return nil
}
