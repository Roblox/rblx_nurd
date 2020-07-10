package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"fmt"
)

type JobData struct {
	JobID       string
	Name        string
	UTicks      float64
	RCPU        float64
	URSS        float64
	UCache      float64
	RMemoryMB   float64
	RdiskMB     float64
	RIOPS       float64
	Namespace   string
	DataCenters string
	CurrentTime string
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
	Count         float64
	Tasks         []Task
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

type JobDesc struct {
	ID          string
	Name        string
	Datacenters []string
	JobSummary  JobSum
}

type JobSum struct {
	Namespace string
}

type Alloc struct {
	ID string
}

func getPromAllocs(clusterAddress, query string, e chan error) map[string]struct{} {
	api := "http://" + clusterAddress + "/api/v1/query?query=" + query
	response, err := http.Get(api) // customize for timeout
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

func getNomadAllocs(clusterAddress, jobID string, e chan error) map[string]struct{} {
	api := "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}

	var allocs []Alloc
	err = json.NewDecoder(response.Body).Decode(&allocs)
	if err != nil {
		e <- err
	}

	m := make(map[string]struct{})
	var Empty struct{}
	for i := range allocs {
		allocID := allocs[i].ID
		m[allocID] = Empty
	}

	return m
}

func getRSS(clusterAddress, metricsAddress, jobID, name string, remainders map[string][]string, e chan error) float64 {
	var rss float64
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22" + name + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	var promStats RawAlloc
	json.NewDecoder(response.Body).Decode(&promStats)
	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			e <- err
		}
		rss += num / 1.049e6
	}

	nomadAllocs := getNomadAllocs(clusterAddress, jobID, e)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_memory_rss_value", e)
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			remainders[allocID] = append(remainders[allocID], "rss")
		}
	}

	return rss
}

func getCache(clusterAddress, metricsAddress, jobID, name string, remainders map[string][]string, e chan error) float64 {
	var cache float64
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22" + name + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	var promStats RawAlloc
	json.NewDecoder(response.Body).Decode(&promStats)
	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			e <- err
		}
		cache += num / 1.049e6
	}

	nomadAllocs := getNomadAllocs(clusterAddress, jobID, e)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_memory_cache_value", e)
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			remainders[allocID] = append(remainders[allocID], "cache")
		}
	}

	return cache
}

func getTicks(clusterAddress, metricsAddress, jobID, name string, remainders map[string][]string, e chan error) float64 {
	var ticks float64
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22" + name + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	var promStats RawAlloc
	json.NewDecoder(response.Body).Decode(&promStats)
	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			e <- err
		}
		ticks += num
	}

	nomadAllocs := getNomadAllocs(clusterAddress, jobID, e)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_cpu_total_ticks_value", e)
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			remainders[allocID] = append(remainders[allocID], "ticks")
		}
	}

	return ticks
}

func getRemainderNomad(clusterAddress string, remainders map[string][]string, e chan error) (float64, float64, float64) {
	var rss, cache, ticks float64

	for allocID, slice := range remainders {
		api := "http://" + clusterAddress + "/v1/client/allocation/" + allocID + "/stats"
		response, err := http.Get(api)
		if err != nil {
			e <- err
		}
		var nomadAlloc NomadAlloc
		json.NewDecoder(response.Body).Decode(&nomadAlloc)

		for _, val := range slice {
			if nomadAlloc.ResourceUsage != (MemCPU{}) {
				resourceUsage := nomadAlloc.ResourceUsage
				memoryStats := resourceUsage.MemoryStats
				cpuStats := resourceUsage.CpuStats
				if val == "rss" {
					rss += memoryStats.RSS / 1.049e6
				} else if val == "cache" {
					cache += memoryStats.Cache / 1.049e6
				} else if val == "ticks" {
					ticks += cpuStats.TotalTicks
				}
			}
		}
	}

	return rss, cache, ticks
}

func aggUsageResources(clusterAddress, metricsAddress, jobID, name string, e chan error) (float64, float64, float64) {
	var rss, ticks, cache float64
	remainders := make(map[string][]string)

	rss = getRSS(clusterAddress, metricsAddress, jobID, name, remainders, e)
	cache = getCache(clusterAddress, metricsAddress, jobID, name, remainders, e)
	ticks = getTicks(clusterAddress, metricsAddress, jobID, name, remainders, e)

	rssRemainder, cacheRemainder, ticksRemainder := getRemainderNomad(clusterAddress, remainders, e)
	rss += rssRemainder
	cache += cacheRemainder
	ticks += ticksRemainder

	return rss, ticks, cache
}

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

func reachCluster(clusterAddress, metricsAddress string, c chan []JobData, e chan error) {
	var jobsClean []JobData

	api := "http://" + clusterAddress + "/v1/jobs"
	response, err := http.Get(api)
	if err != nil {
		e <- err
	}
	var jobs []JobDesc
	json.NewDecoder(response.Body).Decode(&jobs)

	for i := range jobs {
		fmt.Println("Getting job", i, "resources...")
		jobID := jobs[i].ID
		name := jobs[i].Name
		dataCentersSlice := jobs[i].Datacenters
		namespace := jobs[i].JobSummary.Namespace
		rss, ticks, cache := aggUsageResources(clusterAddress, metricsAddress, jobID, name, e)
		CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal := aggReqResources(clusterAddress, jobID, e)

		var dataCenters string
		for i, v := range dataCentersSlice {
			dataCenters += v
			if i != len(dataCentersSlice)-1 {
				dataCenters += ","
			}
		}

		currentTime := time.Now().Format("2006-01-02 15:04:05")
		jobData := JobData{
			jobID,
			name,
			ticks,
			CPUTotal,
			rss,
			cache,
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
