package main

import (
	// "database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	// "strconv"
	"sync"

	// _ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

type JobData struct {
	JobID string
	Data map[string] interface{}
}

func getJobResources(address, jobID string) map[string]interface{} {
	api := "http://" + address + "/v1/job/" + jobID
	response, _ := http.Get(api)
	data, _ := ioutil.ReadAll(response.Body)
	jobJSON := string(data)

	var result map[string]interface{}

	json.Unmarshal([]byte(jobJSON), &result)

	// unpacking JSON
	taskGroups := result["TaskGroups"].([]interface{})[0]
	tasks := taskGroups.(map[string]interface{})["Tasks"].([]interface{})[0]
	resources := tasks.(map[string]interface{})["Resources"].(map[string]interface{})

	return resources
} 

func accessJobs(v string, c chan []JobData) {
	defer wg.Done()
	api := "http://" + v + "/v1/jobs" 
	response, _ := http.Get(api)
	data, _ := ioutil.ReadAll(response.Body)
	sliceOfJsons := string(data)
	keysBody := []byte(sliceOfJsons)
	keys := make([]interface{}, 0)
	json.Unmarshal(keysBody, &keys)
	var jobDataSlice []JobData
		for i := range keys {
			jobID := keys[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["JobID"] // unpack JobID from JSON
			jobDataSlice = append(jobDataSlice, JobData{jobID.(string), getJobResources(v, jobID.(string))})
		}
	c <- jobDataSlice
}

func main() {
	// will be provided 1 address/cluster
	addresses := []string{"***REMOVED***:***REMOVED***", "***REMOVED***:***REMOVED***"} // substitute for config file, server address
	// addresses := []string{}
	buffer := len(addresses)
	c := make(chan []JobData, buffer)
	m := make(map[string]JobData)
	
	for _, v := range addresses {
		wg.Add(1)
		go accessJobs(v, c)
	}

	wg.Wait()
	close(c)
	
	for jobDataSlice := range c {
		for _, v := range jobDataSlice {
			m[v.JobID] = v
		}
	}
	

	i := 0
	for key, val := range m {
		fmt.Println(i, ":", key)
		fmt.Println(val)
		i += 1
	}

	fmt.Println("Complete.")
}