package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
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
	Name          string
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
	DiskMB   float64
	IOPS     float64
}

type JobDesc struct {
	ID          string
	Name        string
	Datacenters []string
	Type        string
	JobSummary  JobSum
}

type JobSum struct {
	Namespace string
}

type Alloc struct {
	ID        string
	TaskGroup string
}

type Error struct {
	FuncName string
	Err	error
}

func getPromAllocs(clusterAddress, query string, e chan Error) map[string]struct{} {
	api := "http://" + clusterAddress + "/api/v1/query?query=" + query
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"getPromAllocs", err}
	}

	var allocs RawAlloc
	err = json.NewDecoder(response.Body).Decode(&allocs)
	if err != nil {
		e <- Error{"getPromAllocs", err}
	}

	m := make(map[string]struct{})
	var empty struct{}
	for _, val := range allocs.Data.Result {
		m[val.Metric.Alloc_id] = empty
	}

	return m
}

func getNomadAllocs(clusterAddress, jobID string, e chan Error) map[string]struct{} {
	api := "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"getNomadAllocs", err}
	}

	var allocs []Alloc
	err = json.NewDecoder(response.Body).Decode(&allocs)
	if err != nil {
		e <- Error{"getNomadAllocs", err}
	}

	m := make(map[string]struct{})
	var empty struct{}
	for _, alloc := range allocs {
		m[alloc.ID] = empty
	}

	return m
}

func getRSS(clusterAddress, metricsAddress, jobID, jobName string, remainders map[string][]string, e chan Error) float64 {
	var rss float64
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22" + jobName + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"getRSS", err}
	}

	var promStats RawAlloc
	err = json.NewDecoder(response.Body).Decode(&promStats)
	if err != nil {
		e <- Error{"getRSS", err}
	}

	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			e <- Error{"getRSS", err}
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

func getCache(clusterAddress, metricsAddress, jobID, jobName string, remainders map[string][]string, e chan Error) float64 {
	var cache float64
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22" + jobName + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"getCache", err}
	}

	var promStats RawAlloc
	err = json.NewDecoder(response.Body).Decode(&promStats)
	if err != nil {
		e <- Error{"getCache", err}
	}

	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			e <- Error{"getCache", err}
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

func getTicks(clusterAddress, metricsAddress, jobID, jobName string, remainders map[string][]string, e chan Error) float64 {
	var ticks float64
	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22" + jobName + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"getTicks", err}
	}

	var promStats RawAlloc
	err = json.NewDecoder(response.Body).Decode(&promStats)
	if err != nil {
		e <- Error{"getTicks", err}
	}

	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			e <- Error{"getTicks", err}
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

func getRemainderNomad(clusterAddress string, remainders map[string][]string, e chan Error) (float64, float64, float64) {
	var rss, cache, ticks float64

	for allocID, slice := range remainders {
		api := "http://" + clusterAddress + "/v1/client/allocation/" + allocID + "/stats"
		response, err := http.Get(api)
		if err != nil {
			e <- Error{"getRemainderNomad", err}
		}

		var nomadAlloc NomadAlloc
		err = json.NewDecoder(response.Body).Decode(&nomadAlloc)
		if err != nil {
			e <- Error{"getRemainderNomad", err}
		}

		for _, val := range slice {
			if nomadAlloc.ResourceUsage != (MemCPU{}) {
				resourceUsage := nomadAlloc.ResourceUsage
				memoryStats := resourceUsage.MemoryStats
				cpuStats := resourceUsage.CpuStats
				switch val {
				case "rss":
					rss += memoryStats.RSS / 1.049e6
				case "cache":
					cache += memoryStats.Cache / 1.049e6
				case "ticks":
					ticks += cpuStats.TotalTicks
				}
			}
		}
	}

	return rss, cache, ticks
}

func aggUsed(clusterAddress, metricsAddress, jobID, jobName string, e chan Error) (float64, float64, float64) {
	remainders := make(map[string][]string)

	rss := getRSS(clusterAddress, metricsAddress, jobID, jobName, remainders, e)
	cache := getCache(clusterAddress, metricsAddress, jobID, jobName, remainders, e)
	ticks := getTicks(clusterAddress, metricsAddress, jobID, jobName, remainders, e)

	rssRemainder, cacheRemainder, ticksRemainder := getRemainderNomad(clusterAddress, remainders, e)
	rss += rssRemainder
	cache += cacheRemainder
	ticks += ticksRemainder

	return rss, ticks, cache
}

func aggRequested(clusterAddress, metricsAddress, jobID, jobType string, e chan Error) (float64, float64, float64, float64) {
	var cpu, memoryMB, diskMB, iops, count float64

	api := "http://" + clusterAddress + "/v1/job/" + jobID
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"aggRequested", err}
	}

	var jobSpec JobSpec
	err = json.NewDecoder(response.Body).Decode(&jobSpec)
	if err != nil {
		e <- Error{"aggRequested", err}
	}

	if jobSpec.TaskGroups == nil {
		return 0, 0, 0, 0
	}

	mapTaskGroupCount := make(map[string]float64)
	if jobType == "system" {
		api = "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
		response, err := http.Get(api)
		if err != nil {
			e <- Error{"aggRequested", err}
		}

		var allocs []Alloc
		err = json.NewDecoder(response.Body).Decode(&allocs)
		for _, alloc := range allocs {
			mapTaskGroupCount[alloc.TaskGroup] += 1
		}
	}

	for _, taskGroup := range jobSpec.TaskGroups {
		switch jobType {
		case "service":
			count = taskGroup.Count
		case "system":
			count = mapTaskGroupCount[taskGroup.Name]
		}

		for _, task := range taskGroup.Tasks {
			resources := task.Resources
			cpu += count * resources.CPU
			memoryMB += count * resources.MemoryMB
			iops += count * resources.IOPS
		}
		diskMB += count * taskGroup.EphemeralDisk.SizeMB
	}

	return cpu, memoryMB, diskMB, iops
}

func reachCluster(clusterAddress, metricsAddress string, c chan []JobData, e chan Error) {
	var jobData []JobData
	var rss, ticks, cache, CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal float64
	api := "http://" + clusterAddress + "/v1/jobs"
	response, err := http.Get(api)
	if err != nil {
		e <- Error{"reachCluster", err}
	}
	if response == nil {
		e <- Error{"reachCluster", errors.New("nil response from Nomad API /v1/jobs")}
		wg.Done()
		return
	}

	var jobs []JobDesc
	err = json.NewDecoder(response.Body).Decode(&jobs)
	if err != nil {
		e <- Error{"reachCluster", err}
	}

	logFile, err := os.OpenFile("nurd.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		e <- Error{"reachCluster", err}
	}
	log.SetOutput(logFile)
	log.SetLevel(log.TraceLevel)

	for _, job := range jobs {
		log.Trace(job.ID)

		if job.Type != "system" && job.Type != "service" {
			continue
		}
		rss, ticks, cache = aggUsed(clusterAddress, metricsAddress, job.ID, job.Name, e)
		CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal = aggRequested(clusterAddress, metricsAddress, job.ID, job.Type, e)

		var dataCenters string
		for i, val := range job.Datacenters {
			dataCenters += val
			if i != len(job.Datacenters)-1 {
				dataCenters += ","
			}
		}

		currentTime := time.Now().Format("2006-01-02 15:04:05")
		jobStruct := JobData{
			job.ID,
			job.Name,
			ticks,
			CPUTotal,
			rss,
			cache,
			memoryMBTotal,
			diskMBTotal,
			IOPSTotal,
			job.JobSummary.Namespace,
			dataCenters,
			currentTime,
		}
		jobData = append(jobData, jobStruct)
	}

	c <- jobData
	wg.Done()
}
