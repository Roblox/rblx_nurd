package main

import (
	"encoding/json"
	"io/ioutil"
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

	data, err := ioutil.ReadFile(path)
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
