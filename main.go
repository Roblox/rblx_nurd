package main

import (
	"database/sql"
	"encoding/json"
	// "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	// "strings"
	"sync"
	
	_ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

// JobData holds namespace, DC,
// resource data for a job
type JobData struct { 
	JobID string
	name string
	uTicks float64
	rCPU float64
	uRSS float64
	pRSS float64
	rMemoryMB float64
	rdiskMB float64
	rIOPS float64
	namespace string
	dataCenters string
	currentTime string
}

func getAllocIDFromProm(address string) map[string]string {
	api := "http://" + address + "/api/v1/query?query=nomad_client_allocs_memory_allocated_value"
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

func getAllocIDFromNomadByJobID(address, jobID string) map[string]string {
	api := "http://" + address + "/v1/job/" + jobID + "/allocations"
	response, _ := http.Get(api)
	data, _ := ioutil.ReadAll(response.Body)

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

// Aggregate total CPU, memory usage for a job
func aggUsageResources(address, jobID, name string, e chan error) (float64, float64, float64) {
	var ticksTotal, rssTotal, cacheTotal, swapTotal, usageTotal, maxUsageTotal, kernelUsageTotal, kernelMaxUsageTotal, rssProm float64
	promAllocs := getAllocIDFromProm("***REMOVED***:***REMOVED***")

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
				cache := memoryStats["Cache"]
				swap := memoryStats["Swap"]
				usage := memoryStats["Usage"]
				maxUsage := memoryStats["MaxUsage"]
				kernelUsage := memoryStats["KernelUsage"]
				kernelMaxUsage := memoryStats["KernelMaxUsage"]
				ticks := cpuStats["TotalTicks"]

				rssTotal += rss.(float64) / 1.049e6
				cacheTotal += cache.(float64) / 1.049e6
				swapTotal += swap.(float64) / 1.049e6
				usageTotal += usage.(float64) / 1.049e6
				maxUsageTotal += maxUsage.(float64) / 1.049e6
				kernelUsageTotal += kernelUsage.(float64) / 1.049e6
				kernelMaxUsageTotal += kernelMaxUsage.(float64) / 1.049e6
				ticksTotal += ticks.(float64)
			}
		}
	}

	nomadAllocs := getAllocIDFromNomadByJobID(address, jobID)

	for key := range nomadAllocs {
		if _, ok := promAllocs[key]; !ok { // if key is not in promAllocs
			clientAllocAPI := "http://" + address + "/v1/client/allocation/" + key + "/stats"
			allocResponse, _ := http.Get(clientAllocAPI)
			allocData, _ := ioutil.ReadAll(allocResponse.Body)
			var allocStats map[string]interface{}
			json.Unmarshal([]byte(string(allocData)), &allocStats)

			if allocStats["ResourceUsage"] != nil {
				resourceUsage := allocStats["ResourceUsage"].(map[string]interface{})
				memoryStats := resourceUsage["MemoryStats"].(map[string]interface{})
				
				rssProm += memoryStats["RSS"].(float64) / 1.049e6
			}

		}
	}
	// Here: sum remaining usage stats from prom
	// rssTotal += sum(prom RSS)
	promAPI := "http://***REMOVED***:***REMOVED***/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22" + name + "%22%7D)%20by%20(job)"
	promResponse, _ := http.Get(promAPI)
	promData, _ := ioutil.ReadAll(promResponse.Body)
	var promStats map[string]interface{}
	json.Unmarshal([]byte(string(promData)), &promStats)
	if len(promStats["data"].(map[string]interface{})["result"].([]interface{})) != 0 {
		num, _ := strconv.ParseFloat(promStats["data"].(map[string]interface{})["result"].([]interface{})[0].(map[string]interface{})["value"].([]interface{})[1].(string), 64)
		rssProm += num / 1.049e6
	}

	return ticksTotal, rssTotal, rssProm
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

// Access a single cluster
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
		name := keys[i].(map[string]interface{})["Name"].(string)
		ticksUsage, rssUsage, rssProm := aggUsageResources(address, jobID, name, e)
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
						name, 
						ticksUsage, 
						CPUTotal, 
						rssUsage,
						rssProm, 
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

// Initialize database
// Configure insert SQL statement
func initDB(nameDB string) (*sql.DB, *sql.Stmt) {
	db, errOpen := sql.Open("sqlite3", nameDB + ".db")

	if errOpen != nil {
		log.Fatal("Error:", errOpen)
	}

	createTable, errCreateTable := db.Prepare(`CREATE TABLE IF NOT EXISTS resources (id INTEGER PRIMARY KEY,
		JobID TEXT,
		uTicks REAL,
		rCPU REAL, 
		uRSS REAL,
		pRSS REAL,
		rMemoryMB REAL,
		rdiskMB REAL,
		rIOPS REAL,
		namespace TEXT,
		dataCenters TEXT,
		date DATETIME)`)
	createTable.Exec()

	if errCreateTable != nil {
		log.Fatal("Error:", errCreateTable)
	}

	insert, errInsert := db.Prepare(`INSERT INTO resources (JobID,
		uTicks, 
		rCPU,
		uRSS,
		pRSS,
		rMemoryMB,
		rdiskMB,
		rIOPS,
		namespace,
		dataCenters,
		date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if errInsert != nil {
		log.Fatal("Error:", errInsert)
	}

	return db, insert
}

func printRowsDB(db *sql.DB) {
	rows, _ := db.Query("SELECT * FROM resources")

	var JobID, namespace, dataCenters, currentTime string
	var uTicks, rCPU, uRSS, rMemoryMB, rdiskMB, rIOPS, pRSS float64
	var id int

	for rows.Next() {
		rows.Scan(&id, &JobID, &uTicks, &rCPU, &uRSS, &pRSS, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currentTime)
		fmt.Println(strconv.Itoa(id) + ": ", JobID, uTicks, rCPU, uRSS, pRSS, rMemoryMB, rdiskMB, rIOPS, namespace, dataCenters, currentTime)
	}
}

func main() {
	addresses, buffer, duration := Config()

	db, insert := initDB("resources")
	
	// While loop for scrape frequency
	for {
		c := make(chan []JobData, buffer)
		e := make(chan error)
		// m := make(map[string]JobData)
	
		begin := time.Now()
	
		// Listen for errors
		go func(e chan error) {
			err := <-e
			log.Fatal("Error: ", err)	
		}(e)
		
		// Goroutines for each cluster address
		for _, address := range addresses {
			wg.Add(1)
			go reachCluster(address, c, e)
		}
	
		wg.Wait()
		close(c)
	
		end := time.Now()
		
		// Insert into db from channel
		for jobDataSlice := range c {
			for _, v := range jobDataSlice {
				insert.Exec(v.JobID,
					v.uTicks,
					v.rCPU,
					v.uRSS,
					v.pRSS,
					v.rMemoryMB,
					v.rdiskMB,
					v.rIOPS,
					v.namespace,
					v.dataCenters,
					v.currentTime)
			}
		}

		printRowsDB(db)
		// this prints rows fo the 
		fmt.Println("Elapsed:", end.Sub(begin))
		time.Sleep(duration)
	}
}