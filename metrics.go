package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "log"
	"net/http"
	"strconv"
)

type JobData struct {
	JobID string
	DC string
}

func getJobIDFromNomad(address string) ([]JobData) {
	api := "http://" + address + "/v1/jobs" 
	response, _ := http.Get(api)	
	data, _ := ioutil.ReadAll(response.Body)
	
	sliceOfJsons := string(data)
	keysBody := []byte(sliceOfJsons)
	keys := make([]interface{}, 0)
	json.Unmarshal(keysBody, &keys)

	s := make([]JobData, 0)
	for i := range keys {
		jobID := keys[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["JobID"].(string) // unpack JobID from JSON
		DC := keys[i].(map[string]interface{})["Datacenters"].([]interface{})[0].(string)
		s = append(s, JobData{jobID, DC})
	}
	
	return s
}

func getAllocIDFromPromByJobID(address, jobID, DC string) (map[string]string) {
	api := "http://" + address + "/api/v1/query?query=nomad_client_allocs_memory_allocated_value%7Bjob%3D%22" + jobID + "%22%2C%20dc%3D%22" + DC + "%22%7D&time=1593023624"
	response, _ := http.Get(api)

	raw, _ := ioutil.ReadAll(response.Body)
	var allocs map[string]interface{}
	json.Unmarshal([]byte(string(raw)), &allocs)

	data := allocs["data"].(map[string]interface{})
	result := data["result"].([]interface{})
	
	// fmt.Println(result[0].(map[string]interface{})["metric"].(map[string]interface{})["job"])
	m := make(map[string]string)
	for _, v := range result {
		alloc_id := v.(map[string]interface{})["metric"].(map[string]interface{})["alloc_id"].(string)
		job := v.(map[string]interface{})["metric"].(map[string]interface{})["job"].(string)
		m[alloc_id] = job
	}
	return m
}

func getAllocIDFromProm(address string) (map[string]string) {
	api := "http://" + address + "/api/v1/query?query=nomad_client_allocs_memory_allocated_value&time=1593475660"
	response, _ := http.Get(api)

	raw, _ := ioutil.ReadAll(response.Body)
	var allocs map[string]interface{}
	json.Unmarshal([]byte(string(raw)), &allocs)

	data := allocs["data"].(map[string]interface{})
	result := data["result"].([]interface{})
	
	// fmt.Println(result[0].(map[string]interface{})["metric"].(map[string]interface{})["job"])
	m := make(map[string]string)
	for _, v := range result {
		alloc_id := v.(map[string]interface{})["metric"].(map[string]interface{})["alloc_id"].(string)
		m[alloc_id] = "value"
	}
	return m
}

func getJobIDFromProm(address string) (map[string]string) {
	api := "http://" + address + "/api/v1/query?query=nomad_client_allocs_memory_allocated_value&time=1593023624"
	response, _ := http.Get(api)

	raw, _ := ioutil.ReadAll(response.Body)
	var allocs map[string]interface{}
	json.Unmarshal([]byte(string(raw)), &allocs)

	data := allocs["data"].(map[string]interface{})
	result := data["result"].([]interface{})
	
	// fmt.Println(result[0].(map[string]interface{})["metric"].(map[string]interface{})["job"])
	// s := make([]string, 2)
	m := make(map[string]string)
	for _, v := range result {
		job := v.(map[string]interface{})["metric"].(map[string]interface{})["job"].(string)
		m[job] = "value"
	}
	return m
}

func getAllocIDFromNomadByJobID(address, jobID string) (map[string]string) {
	api := "http://" + address + "/v1/job/" + jobID + "/allocations"
	response, _ := http.Get(api)
	data, _:= ioutil.ReadAll(response.Body)

	sliceOfAllocs := []byte(string(data))
	keys := make([]interface{}, 0)
	json.Unmarshal(sliceOfAllocs, &keys)

	m := make(map[string]string)

	for i := range keys {
		allocID := keys[i].(map[string]interface{})["ID"].(string)
		m[allocID] = "value"
	}

	return m
}

func main() {
	nomadJobs := getJobIDFromNomad("***REMOVED***:***REMOVED***")
	// promJobs := getJobIDFromProm("***REMOVED***:***REMOVED***")

	promAllocs := getAllocIDFromProm("***REMOVED***:***REMOVED***")
	
	for _, val := range nomadJobs {
		fmt.Println("Job:", val.JobID)
		nomadAllocs := getAllocIDFromNomadByJobID("***REMOVED***:***REMOVED***", val.JobID)
		var rssTotal float64

		for key, _ := range nomadAllocs {
			if _, ok := promAllocs[key]; !ok {
				fmt.Println("NOT IN PROM:", key)
				clientAllocAPI := "http://" + "***REMOVED***:***REMOVED***" + "/v1/client/allocation/" + key + "/stats"
				allocResponse, _ := http.Get(clientAllocAPI)
				allocData, _ := ioutil.ReadAll(allocResponse.Body)
				var allocStats map[string]interface{}
				json.Unmarshal([]byte(string(allocData)), &allocStats)

				if allocStats["ResourceUsage"] != nil {
					resourceUsage := allocStats["ResourceUsage"].(map[string]interface{})
					memoryStats := resourceUsage["MemoryStats"].(map[string]interface{})
					rss := memoryStats["RSS"]
					rssTotal += rss.(float64) / 1.049e6
				}
		
			}
		}
		// Here: sum remaining usage stats from prom
		// rssTotal += sum(prom RSS)
		promAPI := "http://***REMOVED***:***REMOVED***/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22" + val.JobID + "%22%7D)%20by%20(job)&time=1593475660"
		promResponse, _ := http.Get(promAPI)
		promData, _ := ioutil.ReadAll(promResponse.Body)
		var promStats map[string]interface{}
		json.Unmarshal([]byte(string(promData)), &promStats)
		if len(promStats["data"].(map[string]interface{})["result"].([]interface{})) != 0 {
			num, _ := strconv.ParseFloat(promStats["data"].(map[string]interface{})["result"].([]interface{})[0].(map[string]interface{})["value"].([]interface{})[1].(string), 64)
			rssTotal += num / 1.049e6
			fmt.Println("	PROM:", num / 1.049e6)
		}
		fmt.Println("	RSS:", rssTotal)
	}
}