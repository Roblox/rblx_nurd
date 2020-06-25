package main

import (
	// "database/sql"
	"encoding/json"
	// "errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	// "strconv"
	"sync"
	// "time"
	// _ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

type JobData struct { 
	JobID string
	ticksUsed float64
	CPURequested float64
	RSSUsed float64
	memoryMBRequested float64
	diskMB float64
	IOPS float64
}

func aggUsageResources(address, jobID string) (float64, float64) {
	var totalTicksUsageTotal, rssUsageTotal float64

	jobsAPI := "http://" + address + "/v1/job/" + jobID + "/allocations"
	response, _ := http.Get(jobsAPI)
	data, _ := ioutil.ReadAll(response.Body)
	sliceOfAllocs := []byte(string(data))

	keys := make([]interface{}, 0)
	json.Unmarshal(sliceOfAllocs, &keys)

	// prints out alloc ids for a specified job
	for i := range keys {
		// fmt.Println("allocID")
		allocID := keys[i].(map[string]interface{})["ID"].(string)
		clientStatus := keys[i].(map[string]interface{})["ClientStatus"].(string)

		if clientStatus != "lost" {
			clientAllocAPI := "http://" + address + "/v1/client/allocation/" + allocID + "/stats"
			allocResponse, _ := http.Get(clientAllocAPI)
			allocData, _ := ioutil.ReadAll(allocResponse.Body)
	
			var allocStats map[string]interface{}
			json.Unmarshal([]byte(string(allocData)), &allocStats)

			if allocStats["ResourceUsage"] != nil {
				// fmt.Println("ResourceUsage", jobID)
				// fmt.Println(allocStats["ResourceUsage"])
				resourceUsage := allocStats["ResourceUsage"].(map[string]interface{})
				// fmt.Println("MemoryStats")
				memoryStats := resourceUsage["MemoryStats"].(map[string]interface{})
				// fmt.Println("CpuStats")
				cpuStats := resourceUsage["CpuStats"].(map[string]interface{})
		
				rss := memoryStats["RSS"]

				totalTicks := cpuStats["TotalTicks"]
		
				rssUsageTotal += rss.(float64)
				totalTicksUsageTotal += totalTicks.(float64)
			}
		}
	}

	// fmt.Println("Total RSS Usage:", rssUsageTotal)
	// fmt.Println("Total Ticks Usage:", totalTicksUsageTotal)
	return totalTicksUsageTotal, rssUsageTotal / 1e6

}
 
func aggReqResources(address, jobID string, e chan error) (float64, float64, float64, float64) {
	var CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal float64

	api := "http://" + address + "/v1/job/" + jobID
	response, errHttp := http.Get(api)

	// errHttp = errors.New("HTTP ERROR - aggResources(address, jobID, e)")
	if errHttp != nil {
		e <- errHttp
	}

	data, errIoutil := ioutil.ReadAll(response.Body)
	
	// errIoutil = errors.New("IOUTIL ERROR - aggResources(address, jobID, e)")
	if errIoutil != nil {
		e <- errIoutil
	}

	jobJSON := string(data)
	var jobSpec map[string]interface{}
	json.Unmarshal([]byte(jobJSON), &jobSpec)

	taskGroups := jobSpec["TaskGroups"].([]interface{})
	for _, taskGroup := range taskGroups {
		count := taskGroup.(map[string]interface{})["Count"].(float64)
		tasks := taskGroup.(map[string]interface{})["Tasks"].([]interface{})

		for _, task := range tasks {
			resources := task.(map[string]interface{})["Resources"].(map[string]interface{})
			CPUTotal += count * resources["CPU"].(float64)
			memoryMBTotal += count * resources["MemoryMB"].(float64)
			diskMBTotal += count * resources["DiskMB"].(float64)
			IOPSTotal += count * resources["IOPS"].(float64)
		}
	}
	return CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal
} 

func reachCluster(address string, c chan []JobData, e chan error) {
	api := "http://" + address + "/v1/jobs" 
	response, errHttp := http.Get(api)

	// errHttp = errors.New("HTTP ERROR - reachCluster(address, c, e)")
	if errHttp != nil {
		e <- errHttp
	}
	
	data, errIoutil := ioutil.ReadAll(response.Body)
	
	// errIoutil = errors.New("IOUTIL ERROR - reachCluster(address, c, e)")
	if errIoutil != nil {
		e <- errIoutil
	}
	
	sliceOfJsons := string(data)
	keysBody := []byte(sliceOfJsons)
	keys := make([]interface{}, 0)

	json.Unmarshal(keysBody, &keys)

	var jobDataSlice []JobData

	for i := range keys {
		jobID := keys[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["JobID"].(string) // unpack JobID from JSON
		CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal := aggReqResources(address, jobID, e)
		ticksUsage, rssUsage := aggUsageResources(address, jobID)

		jobData := JobData{jobID, ticksUsage, CPUTotal, rssUsage, memoryMBTotal, diskMBTotal, IOPSTotal}

		jobDataSlice = append(jobDataSlice, jobData)
	}

	c <- jobDataSlice

	wg.Done()
}

func main() {
	// will be provided 1 address/cluster
	addresses := []string{"***REMOVED***:***REMOVED***", "***REMOVED***:***REMOVED***"} // substitute for config file, server address
	buffer := len(addresses)
	c := make(chan []JobData, buffer)
	e := make(chan error)
	m := make(map[string]JobData)

	go func(e chan error) {
		err := <-e
		log.Fatal("Error: ", err)	
	}(e)
	
	for _, address := range addresses {
		wg.Add(1)
		go reachCluster(address, c, e)
	}

	wg.Wait()
	close(c)
	
	for jobDataSlice := range c {
		for _, v := range jobDataSlice {
			m[v.JobID] = v
		}
	}

	// may not have to use hash table for aggregation, duplicate filter
	// since aggregation is done in aggResources()
	// and duplicates should be filtered out by separate cluster addresses
	i := 0
	for _, val := range m {
		fmt.Println(i, ":", val)
		i += 1
	}

	fmt.Println("Complete.")
}