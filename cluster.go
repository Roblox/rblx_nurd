package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func getPromAllocs(clusterAddress, query string) map[string]struct{} {
	m := make(map[string]struct{})

	log.SetReportCaller(true)

	api := "http://" + clusterAddress + "/api/v1/query?query=" + query
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		return nil
	}
	defer response.Body.Close()

	var allocs RawAlloc
	err = json.NewDecoder(response.Body).Decode(&allocs)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return nil
	}

	var empty struct{}
	for _, val := range allocs.Data.Result {
		m[val.Metric.Alloc_id] = empty
	}

	return m
}

func getNomadAllocs(clusterAddress, jobID string) map[string]struct{} {
	m := make(map[string]struct{})

	log.SetReportCaller(true)

	api := "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		return nil
	}
	defer response.Body.Close()

	var allocs []Alloc
	err = json.NewDecoder(response.Body).Decode(&allocs)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return nil
	}

	var empty struct{}
	for _, alloc := range allocs {
		m[alloc.ID] = empty
	}

	return m
}

func getRSS(clusterAddress, metricsAddress, jobID, jobName string, remainders map[string][]string) float64 {
	var rss float64

	log.SetReportCaller(true)

	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22" + jobName + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		nomadAllocs := getNomadAllocs(clusterAddress, jobID)
		for allocID := range nomadAllocs {
			remainders[allocID] = append(remainders[allocID], "rss")
		}
		return rss
	}
	defer response.Body.Close()

	var promStats RawAlloc
	err = json.NewDecoder(response.Body).Decode(&promStats)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return rss
	}

	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			log.Error(fmt.Sprintf("Error in parsing float: %v", err))
			return rss
		}
		rss += num / 1.049e6
	}

	nomadAllocs := getNomadAllocs(clusterAddress, jobID)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_memory_rss_value")
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			remainders[allocID] = append(remainders[allocID], "rss")
		}
	}

	return rss
}

func getCache(clusterAddress, metricsAddress, jobID, jobName string, remainders map[string][]string) float64 {
	var cache float64

	log.SetReportCaller(true)

	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22" + jobName + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		nomadAllocs := getNomadAllocs(clusterAddress, jobID)
		for allocID := range nomadAllocs {
			remainders[allocID] = append(remainders[allocID], "cache")
		}
		return cache
	}
	defer response.Body.Close()

	var promStats RawAlloc
	err = json.NewDecoder(response.Body).Decode(&promStats)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return cache
	}

	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			log.Error(fmt.Sprintf("Error in parsing float: %v", err))
			return cache
		}
		cache += num / 1.049e6
	}

	nomadAllocs := getNomadAllocs(clusterAddress, jobID)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_memory_cache_value")
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			remainders[allocID] = append(remainders[allocID], "cache")
		}
	}

	return cache
}

func getTicks(clusterAddress, metricsAddress, jobID, jobName string, remainders map[string][]string) float64 {
	var ticks float64

	log.SetReportCaller(true)

	api := "http://" + metricsAddress + "/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22" + jobName + "%22%7D)%20by%20(job)"
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		nomadAllocs := getNomadAllocs(clusterAddress, jobID)
		for allocID := range nomadAllocs {
			remainders[allocID] = append(remainders[allocID], "ticks")
		}
		return ticks
	}
	defer response.Body.Close()

	var promStats RawAlloc
	err = json.NewDecoder(response.Body).Decode(&promStats)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return ticks
	}

	if len(promStats.Data.Result) != 0 {
		num, err := strconv.ParseFloat(promStats.Data.Result[0].Value[1].(string), 64)
		if err != nil {
			log.Error(fmt.Sprintf("Error in parsing float: %v", err))
			return ticks
		}
		ticks += num
	}

	nomadAllocs := getNomadAllocs(clusterAddress, jobID)
	promAllocs := getPromAllocs(metricsAddress, "nomad_client_allocs_cpu_total_ticks_value")
	for allocID := range nomadAllocs {
		if _, ok := promAllocs[allocID]; !ok {
			remainders[allocID] = append(remainders[allocID], "ticks")
		}
	}

	return ticks
}

func getRemainderNomad(clusterAddress string, remainders map[string][]string) (float64, float64, float64) {
	var rss, cache, ticks float64

	log.SetReportCaller(true)

	for allocID, slice := range remainders {
		api := "http://" + clusterAddress + "/v1/client/allocation/" + allocID + "/stats"
		response, err := http.Get(api)
		if err != nil {
			log.Error(fmt.Sprintf("Error in getting API response: %v", err))
			return rss, cache, ticks
		}
		defer response.Body.Close()

		var nomadAlloc NomadAlloc
		err = json.NewDecoder(response.Body).Decode(&nomadAlloc)
		if err != nil {
			log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
			return rss, cache, ticks
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

func aggUsed(clusterAddress, metricsAddress, jobID, jobName string) (float64, float64, float64) {
	remainders := make(map[string][]string)

	rss := getRSS(clusterAddress, metricsAddress, jobID, jobName, remainders)
	cache := getCache(clusterAddress, metricsAddress, jobID, jobName, remainders)
	ticks := getTicks(clusterAddress, metricsAddress, jobID, jobName, remainders)

	rssRemainder, cacheRemainder, ticksRemainder := getRemainderNomad(clusterAddress, remainders)
	rss += rssRemainder
	cache += cacheRemainder
	ticks += ticksRemainder

	return rss, ticks, cache
}

func aggRequested(clusterAddress, metricsAddress, jobID, jobType string) (float64, float64, float64, float64) {
	var cpu, memoryMB, diskMB, iops, count float64

	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)

	api := "http://" + clusterAddress + "/v1/job/" + jobID
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		return cpu, memoryMB, diskMB, iops
	}
	defer response.Body.Close()

	var jobSpec JobSpec
	err = json.NewDecoder(response.Body).Decode(&jobSpec)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return cpu, memoryMB, diskMB, iops
	}

	if jobSpec.TaskGroups == nil {
		return cpu, memoryMB, diskMB, iops
	}

	mapTaskGroupCount := make(map[string]float64)
	if jobType == "system" {
		api = "http://" + clusterAddress + "/v1/job/" + jobID + "/allocations"
		response, err := http.Get(api)
		if err != nil {
			log.Error(fmt.Sprintf("Error in getting API response: %v", err))
			return cpu, memoryMB, diskMB, iops
		}
		defer response.Body.Close()

		var allocs []Alloc
		err = json.NewDecoder(response.Body).Decode(&allocs)
		if err != nil {
			log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
			return cpu, memoryMB, diskMB, iops
		}

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

func reachCluster(clusterAddress, metricsAddress string, c chan<- []JobData) {
	var jobData []JobData
	var rss, ticks, cache, CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal float64

	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)

	api := "http://" + clusterAddress + "/v1/jobs"
	response, err := http.Get(api)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting API response: %v", err))
		wg.Done()
		return
	}
	defer response.Body.Close()

	var jobs []JobDesc
	err = json.NewDecoder(response.Body).Decode(&jobs)
	if err != nil {
		log.Error(fmt.Sprintf("Error in decoding JSON: %v", err))
		return
	}

	for _, job := range jobs {
		log.Trace(job.ID)

		if job.Type != "system" && job.Type != "service" {
			continue
		}
		rss, ticks, cache = aggUsed(clusterAddress, metricsAddress, job.ID, job.Name)
		CPUTotal, memoryMBTotal, diskMBTotal, IOPSTotal = aggRequested(clusterAddress, metricsAddress, job.ID, job.Type)

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
