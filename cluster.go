package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"fmt"
)

type JobData struct {
	JobID       string
	name        string
	uTicks      float64
	pTicks      float64
	rCPU        float64
	uRSS        float64
	pRSS        float64
	rMemoryMB   float64
	rdiskMB     float64
	rIOPS       float64
	namespace   string
	dataCenters string
	currentTime string
}

type RawAlloc struct {
	Status string
	Data   DataMap
}

type DataMap struct {
	ResultType string
	Result     []MetVal
}

type MetVal struct {
	Metric MetricType
	Value  []interface{}
}

type MetricType struct {
	Alloc_id string
}

type NomadAlloc struct {
	ResourceUsage MemCPU
}

type MemCPU struct {
	MemoryStats Memory
	CpuStats    CPU
}

type Memory struct {
	RSS            float64
	Cache          float64
	Swap           float64
	Usage          float64
	MaxUsage       float64
	KernelUsage    float64
	KernelMaxUsage float64
}

type CPU struct {
	TotalTicks float64
}

type JobSpec struct {
	TaskGroups []TaskGroup
}

type TaskGroup struct {
	Count float64
	Tasks []Task
	EphemeralDisk Disk
}

type Task struct {
	Resources Resource
}

type Disk struct {
	SizeMB float64
}

type Resource struct {
	CPU      float64
	MemoryMB float64
	IOPS     float64
}

func getPromAllocs(clusterAddress, query string, e chan error) map[string]struct{} {
	api := "http://" + clusterAddress + "/api/v1/query?query=" + query //nomad_client_allocs_memory_rss_value
	response, err := http.Get(api)                                     // customize for timeout
	if err != nil {
		e <- err
	}

	var allocs RawAlloc
	err = json.NewDecoder(response.Body).Decode(&allocs)
	if err != nil {
		e <- err
	}

	result := allocs.Data.Result
	m := make(map[string]struct{})
	var Empty struct{}
	for _, v := range result {
		alloc_id := v.Metric.Alloc_id
		m[alloc_id] = Empty
	}

	return m
}

func getNomadAllocs(clusterAddress, jobID string) map[string]string {
	api := "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
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

func getRSS(clusterAddress, metricsAddress, jobID, name string, e chan error) float64 {
	var rss float64

	// Sum RSS stats from Prometheus
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22" + name + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	var promStats RawAlloc
	json.NewDecoder(response.Body).Decode(&promStats)
	if len(promStats.Data.Result) != 0 {
		num, _ := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		rss += num / 1.049e6
	}

	// Get remaining data from Nomad
	nomadAllocs := getNomadAllocs(clusterAddress, jobID)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_memory_rss_value", e)
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			api := "http://" + clusterAddress + "/v1/client/allocation/" + allocID + "/stats"
			response, err := http.Get(api)
			if err != nil {
				e <- err
			}
			var nomadAlloc NomadAlloc
			json.NewDecoder(response.Body).Decode(&nomadAlloc)
			if nomadAlloc.ResourceUsage != (MemCPU{}) {
				resourceUsage := nomadAlloc.ResourceUsage
				memoryStats := resourceUsage.MemoryStats
				rss += memoryStats.RSS / 1.049e6
			}
		}
	}

	return rss
}

// Aggregate total CPU, memory usage for a job
func aggUsageResources(clusterAddress, metricsAddress, jobID, name string, e chan error) (float64, float64, float64, float64) {
	var ticksTotal, rssTotal, cacheTotal, swapTotal, usageTotal, maxUsageTotal, kernelUsageTotal, kernelMaxUsageTotal, rssProm, ticksProm float64

	rssProm = getRSS(clusterAddress, metricsAddress, jobID, name, e)

	api := "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	allocs := make([]interface{}, 0)
	json.NewDecoder(response.Body).Decode(&allocs)
	for i := range allocs {
		allocID := allocs[i].(map[string]interface{})["ID"].(string)
		clientStatus := allocs[i].(map[string]interface{})["ClientStatus"].(string)

		if clientStatus != "lost" {
			clientAllocAPI := "http://" + clusterAddress + "/v1/client/allocation/" + allocID + "/stats"
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

	return ticksTotal, rssTotal, rssProm, ticksProm
}

// Aggregate resources requested by a job
func aggReqResources(clusterAddress, jobID string, e chan error) (float64, float64, float64, float64) {
	var CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal float64

	api := "http://" + clusterAddress + "/v1/job/" + jobID
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	var jobSpec JobSpec
	json.NewDecoder(response.Body).Decode(&jobSpec)
	if jobSpec.TaskGroups == nil {
		fmt.Println("TASKGROUPS NIL\nJOB:", jobID)
		return 0, 0, 0, 0
	}
	taskGroups := jobSpec.TaskGroups
	for _, taskGroup := range taskGroups {
		count := taskGroup.Count
		tasks := taskGroup.Tasks
		ephemeralDisk := taskGroup.EphemeralDisk.SizeMB

		for _, task := range tasks {
			resources := task.Resources
			CPUTotal += count * resources.CPU
			memoryMBTotal += count * resources.MemoryMB
			IOPSTotal += count * resources.IOPS
		}
		diskMBTotal += count * ephemeralDisk
	}

	return CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal
}

// Access a single cluster
func reachCluster(clusterAddress, metricsAddress string, c chan []JobData, e chan error) {
	api := "http://" + clusterAddress + "/v1/jobs"
	response, err := http.Get(api)
	// errHttp = errors.New("HTTP ERROR - reachCluster(address, c, e)")
	if err != nil {
		e <- err
	}

	jobsRaw := make([]interface{}, 0)
	json.NewDecoder(response.Body).Decode(&jobsRaw)
	var jobsClean []JobData

	for i := range jobsRaw {
		jobID := jobsRaw[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["JobID"].(string) // unpack JobID from JSON
		name := jobsRaw[i].(map[string]interface{})["Name"].(string)
		ticksUsage, rssUsage, rssProm, ticksProm := aggUsageResources(clusterAddress, metricsAddress, jobID, name, e)
		CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal := aggReqResources(clusterAddress, jobID, e)
		namespace := jobsRaw[i].(map[string]interface{})["JobSummary"].(map[string]interface{})["Namespace"].(string)
		dataCentersSlice := jobsRaw[i].(map[string]interface{})["Datacenters"].([]interface{})

		var dataCenters string

		for i, v := range dataCentersSlice {
			dataCenters += v.(string)
			if i != len(dataCentersSlice)-1 {
				dataCenters += " "
			}
		}

		currentTime := time.Now().Format("2006-01-02 15:04:05")
		jobData := JobData{
			jobID,
			name,
			ticksUsage,
			ticksProm,
			CPUTotal,
			rssUsage,
			rssProm,
			memoryMBTotal,
			diskMBTotal,
			IOPSTotal,
			namespace,
			dataCenters,
			currentTime}
		jobsClean = append(jobsClean, jobData)
	}

	c <- jobsClean

	wg.Done()
}
