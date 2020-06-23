package main

import (
	// "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

func getJobResources(address, jobID string, e chan error) map[string]interface{} {
	api := "http://" + address + "/v1/job/" + jobID
	response, errHttp := http.Get(api)

	// errHttp = errors.New("HTTP ERROR - getJobResources(address, jobID, e)")
	if errHttp != nil {
		e <- errHttp
	}

	data, errIoutil := ioutil.ReadAll(response.Body)

	// errIoutil = errors.New("IOUTIL ERROR - getJobResources(address, jobID, e)")
	if errIoutil != nil {
		e <- errIoutil
	}

	jobJSON := string(data)

	var result map[string]interface{}

	json.Unmarshal([]byte(jobJSON), &result)

	// unpacking JSON
	taskGroups := result["TaskGroups"].([]interface{})[0]
	tasks := taskGroups.(map[string]interface{})["Tasks"].([]interface{})[0]
	resources := tasks.(map[string]interface{})["Resources"].(map[string]interface{})

	return resources
} 

func accessJobs(v string, c chan []JobData, e chan error) {
	api := "http://" + v + "/v1/jobs" 
	response, errHttp := http.Get(api)

	// errHttp = errors.New("HTTP ERROR - accessJobs(v, c, e)")
	if errHttp != nil {
		e <- errHttp
	}
	
	data, errIoutil := ioutil.ReadAll(response.Body)
	
	// errIoutil = errors.New("IOUTIL ERROR - accessJobs(v, c, e)")
	if errIoutil != nil {
		e <- errIoutil
	}
	
	sliceOfJsons := string(data)
	keysBody := []byte(sliceOfJsons)
	keys := make([]interface{}, 0)
	json.Unmarshal(keysBody, &keys)

	var jobDataSlice []JobData

	for i := range keys {
		jobID := keys[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["JobID"] // unpack JobID from JSON
		jobDataSlice = append(jobDataSlice, JobData{jobID.(string), getJobResources(v, jobID.(string), e)})
	}
	c <- jobDataSlice
	wg.Done()
}

func main() {
	// will be provided 1 address/cluster
	addresses := []string{"***REMOVED***:***REMOVED***", "***REMOVED***:***REMOVED***"} // substitute for config file, server address
	// addresses := []string{}
	buffer := len(addresses)
	c := make(chan []JobData, buffer)
	e := make(chan error)
	m := make(map[string]JobData)

	go func(e chan error) {
		err := <-e
		// close(e)
		log.Fatal("Error: ", err)	
	}(e)
	
	for _, v := range addresses {
		wg.Add(1)
		go accessJobs(v, c, e)
	}

	wg.Wait()
	close(c)

	// select {
	// 	case <-c:
	// 		break
	// 	case err := <-e:
	// 		close(e)
	// 		fmt.Println("Error: ", err)
	// 		log.Fatal("Error: ", err)
	// }
	
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