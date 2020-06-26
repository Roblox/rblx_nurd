package main

import (
	"database/sql"
	"encoding/json"
	// "errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	// "strings"
	"sync"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

// JobData holds namespace, DC,
// resource data for a job
type JobData struct { 
	JobID string
	uTicks float64
	rCPU float64
	uRSS float64
	rMemoryMB float64
	rdiskMB float64
	rIOPS float64
	namespace string
	dataCenters string
	currentTime string
}

// Aggregate total CPU, memory usage for a job
func aggUsageResources(address, jobID string, e chan error) (float64, float64) {
	var ticksTotal, rssTotal float64

	api := "http://" + address + "/v1/job/" + jobID + "/allocations"
	response, errHttp := http.Get(api)
	// errHttp = errors.New("HTTP ERROR - aggUsageResources(address, jobID, e)")
	if errHttp != nil {
		e <- errHttp
	}

	data, errIoutil := ioutil.ReadAll(response.Body)
	// errIoutil = errors.New("IOUTIL ERROR - aggUsageResources(address, jobID, e)")
	if errIoutil != nil {
		e <- errIoutil
	}

	sliceOfAllocs := []byte(string(data))
	keys := make([]interface{}, 0)
	json.Unmarshal(sliceOfAllocs, &keys)

	for i := range keys {
		allocID := keys[i].(map[string]interface{})["ID"].(string)
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
				ticks := cpuStats["TotalTicks"]
				rssTotal += rss.(float64) / 1e6
				ticksTotal += ticks.(float64)
			}
		}
	}

	return ticksTotal, rssTotal 
}

// Aggregate resources requested by a job
func aggReqResources(address, jobID string, e chan error) (float64, float64, float64, float64) {
	var CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal float64

	api := "http://" + address + "/v1/job/" + jobID
	response, errHttp := http.Get(api)
	// errHttp = errors.New("HTTP ERROR - aggReqResources(address, jobID, e)")
	if errHttp != nil {
		e <- errHttp
	}

	data, errIoutil := ioutil.ReadAll(response.Body)
	// errIoutil = errors.New("IOUTIL ERROR - aggReqResources(address, jobID, e)")
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
		ticksUsage, rssUsage := aggUsageResources(address, jobID, e)
		CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal := aggReqResources(address, jobID, e)
		namespace := keys[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["Namespace"].(string)
		dataCentersSlice := keys[i].(map[string]interface{})["Datacenters"].([]interface{})
		var dataCenters string
		for i, v := range dataCentersSlice {
			dataCenters += v.(string)
			if i != len(dataCentersSlice) - 1 {
				dataCenters += " "
			}
		}
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		jobData := JobData{
						jobID, 
						ticksUsage, 
						CPUTotal, 
						rssUsage, 
						memoryMBTotal, 
						diskMBTotal, 
						IOPSTotal,
						namespace,
						dataCenters,
						currentTime}
		jobDataSlice = append(jobDataSlice, jobData)
	}

	c <- jobDataSlice

	wg.Done()
}

func main() {
	// substitute for config file, server address
	addresses := []string{"***REMOVED***:***REMOVED***", "***REMOVED***:***REMOVED***"}
	buffer := len(addresses)
	duration, _ := time.ParseDuration("1m")

	// configure database
	db, _ := sql.Open("sqlite3", "resources.db")
	createTable, _ := db.Prepare(`CREATE TABLE IF NOT EXISTS resources (id INTEGER PRIMARY KEY,
		JobID TEXT,
		uTicks REAL,
		rCPU REAL, 
		uRSS REAL,
		rMemoryMB REAL,
		rdiskMB REAL,
		rIOPS REAL,
		namespace TEXT,
		dataCenters TEXT,
		date DATETIME)`)
	createTable.Exec()

	for {
		c := make(chan []JobData, buffer)
		e := make(chan error)
		m := make(map[string]JobData)
	
		begin := time.Now()
	
		// Listen for errors
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
	
		end := time.Now()
		
		// channel to hash map
		for jobDataSlice := range c {
			for _, v := range jobDataSlice {
				m[v.JobID] = v
			}
		}
	
		// may not have to use hash table for aggregation, duplicate filter
		// since aggregation is done in aggResources()
		// and duplicates should be filtered out by separate cluster addresses
		// i := 0
		insert, errPrepare := db.Prepare(`INSERT INTO resources (JobID,
			uTicks, 
			rCPU,
			uRSS,
			rMemoryMB,
			rdiskMB,
			rIOPS,
			namespace,
			dataCenters,
			date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

		if errPrepare != nil {
			log.Fatal("Error:", errPrepare)
		}

		for _, val := range m {
			// fmt.Println(val.uTicks)
			insert.Exec(val.JobID,
						val.uTicks,
						val.rCPU,
						val.uRSS,
						val.rMemoryMB,
						val.rdiskMB,
						val.rIOPS,
						val.namespace,
						val.dataCenters,
						val.currentTime)

						// fmt.Println(i, ":", val)
						// i += 1
		}

		rows, _ := db.Query("SELECT * FROM resources")

		var JobID, namespace, dataCenters, currentTime string
		var uTicks, rCPU, uRSS, rMemoryMB, rdiskMB, rIOPS float64
		var id int

		for rows.Next() {
			// rows.Scan(&id, &JobID, &uTicks, &rCPU, &uRSS, &rMemoryMB, &rdiskMB, &rIOPS, &namespace)
			// fmt.Println(strconv.Itoa(id) + ":", JobID, uTicks, rCPU, uRSS, rMemoryMB, rdiskMB, rIOPS, namespace)
			rows.Scan(&id, &JobID, &uTicks, &rCPU, &uRSS, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currentTime)
			fmt.Println(strconv.Itoa(id) + ": ", JobID, uTicks, rCPU, uRSS, rMemoryMB, rdiskMB, rIOPS, namespace, dataCenters, currentTime)
		}
		fmt.Println("Elapsed:", end.Sub(begin))
		time.Sleep(duration)
	}
}