/*
Copyright 2020 Roblox Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

	
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
