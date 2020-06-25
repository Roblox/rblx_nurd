package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

)

func aggUsageResources(address, jobID string) (float64, float64) {
	var rssUsageTotal, totalTicksUsageTotal float64

	jobsAPI := "http://" + address + "/v1/job/" + jobID + "/allocations"
	response, _ := http.Get(jobsAPI)
	data, _ := ioutil.ReadAll(response.Body)
	sliceOfAllocs := []byte(string(data))

	keys := make([]interface{}, 0)
	json.Unmarshal(sliceOfAllocs, &keys)

	// prints out alloc ids for a specified job
	for i := range keys {
		allocID := keys[i].(map[string]interface{})["ID"].(string)
		fmt.Println("allocID:", allocID)
		clientStatus := keys[i].(map[string]interface{})["ClientStatus"].(string)
		if clientStatus != "lost" {
			clientAllocAPI := "http://" + address + "/v1/client/allocation/" + allocID + "/stats"
			allocResponse, _ := http.Get(clientAllocAPI)
			allocData, _ := ioutil.ReadAll(allocResponse.Body)
	
			var allocStats map[string]interface{}
			json.Unmarshal([]byte(string(allocData)), &allocStats)
	
			if allocStats["ResourceUsage"] != nil {
				resourceUsage := allocStats["ResourceUsage"].(map[string]interface{})
				memoryStats := resourceUsage["MemoryStats"].(map[string]interface{})
				cpuStats := resourceUsage["CpuStats"].(map[string]interface{})
		
				rss := memoryStats["RSS"]
				cache := memoryStats["Cache"]
				swap := memoryStats["Swap"]
		
				userMode := cpuStats["UserMode"]
				totalTicks := cpuStats["TotalTicks"]
		
				fmt.Println(allocID, "\n--- Memory Stats")
				fmt.Println("	RSS:", rss)
				fmt.Println("	Cache:", cache)
				fmt.Println("	Swap:", swap)
		
				fmt.Println("--- CPU Stats")
				fmt.Println("	User Mode:", userMode)
				fmt.Println("	Total Ticks:", totalTicks, "\n")
		
				rssUsageTotal += rss.(float64)
				totalTicksUsageTotal += totalTicks.(float64)
			}
		}
	}

	fmt.Println("Total Ticks Usage:", totalTicksUsageTotal)
	fmt.Println("Total RSS Usage:", rssUsageTotal)
	return totalTicksUsageTotal, rssUsageTotal

}

func main() {
	address := "***REMOVED***:***REMOVED***"
	jobID := "cas-eaerendil-game-sorts"
	aggUsageResources(address, jobID)
}