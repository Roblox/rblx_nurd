package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

)

func aggUsageResources(address, jobID string) {
	jobsAPI := "http://" + address + "/v1/job/" + jobID + "/allocations"
	response, _ := http.Get(jobsAPI)
	data, _ := ioutil.ReadAll(response.Body)
	sliceOfAllocs := []byte(string(data))

	keys := make([]interface{}, 0)
	json.Unmarshal(sliceOfAllocs, &keys)

	// prints out alloc ids for a specified job
	for i := range keys {
		allocID := keys[i].(map[string]interface{})["ID"].(string)
		clientAllocAPI := "http://" + address + "/v1/client/allocation/" + allocID + "/stats"
		allocResponse, _ := http.Get(clientAllocAPI)
		allocData, _ := ioutil.ReadAll(allocResponse.Body)

		var allocStats map[string]interface{}
		json.Unmarshal([]byte(string(allocData)), &allocStats)

		fmt.Println(allocID)
		fmt.Println(allocStats["ResourceUsage"], "\n")
	}

}

func main() {
	address := "***REMOVED***:***REMOVED***"
	jobID := "grafana-cloudsvcs-alpha"
	aggUsageResources(address, jobID)
}