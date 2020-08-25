/*
Copyright 2020 Roblox Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"io/ioutil"
	"testing"

	"github.com/jarcoal/httpmock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetVMAllocs(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://goodAddress/api/v1/query?query=query1",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
					
					]
				}
			}`,
		),
	)
	expectedVMAllocs := map[string]struct{}{}
	actualVMAllocs := getVMAllocs("goodAddress", "query1")
	assert.Empty(t, actualVMAllocs)
	assert.Equal(t, expectedVMAllocs, actualVMAllocs)

	httpmock.RegisterResponder("GET", "http://goodAddress/api/v1/query?query=query2",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	expectedVMAllocs = map[string]struct{}{
		"alloc_id1": {},
		"alloc_id2": {},
	}
	actualVMAllocs = getVMAllocs("goodAddress", "query2")
	assert.NotNil(t, actualVMAllocs)
	assert.Equal(t, expectedVMAllocs, actualVMAllocs)

	httpmock.RegisterResponder("GET", "http://goodAddress/api/v1/query?query=query3",
		httpmock.NewStringResponder(200, `
			{
				invalid JSON
			}`,
		),
	)
	expectedVMAllocs = nil
	actualVMAllocs = getVMAllocs("goodAddress", "query3")
	assert.Empty(t, actualVMAllocs)
	assert.Equal(t, expectedVMAllocs, actualVMAllocs)

	expectedVMAllocs = nil
	actualVMAllocs = getVMAllocs("goodAddress", "badQuery")
	assert.Empty(t, actualVMAllocs)
	assert.Equal(t, expectedVMAllocs, actualVMAllocs)

	expectedVMAllocs = nil
	actualVMAllocs = getVMAllocs("badAddress", "query2")
	assert.Empty(t, actualVMAllocs)
	assert.Equal(t, expectedVMAllocs, actualVMAllocs)
}

func TestGetNomadAllocs(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://goodAddress/v1/job/job1/allocations",
		httpmock.NewStringResponder(200, `
			[
			]`,
		),
	)
	expectedNomadAllocs := map[string]struct{}{}
	actualNomadAllocs := getNomadAllocs("goodAddress", "job1")
	assert.Empty(t, actualNomadAllocs)
	assert.Equal(t, expectedNomadAllocs, actualNomadAllocs)

	httpmock.RegisterResponder("GET", "http://goodAddress/v1/job/job2/allocations",
		httpmock.NewStringResponder(200, `
			[
				{
					"ID": "ID1"
				},
				{
					"ID": "ID2"
				}
			]`,
		),
	)
	expectedNomadAllocs = map[string]struct{}{
		"ID1": {},
		"ID2": {},
	}
	actualNomadAllocs = getNomadAllocs("goodAddress", "job2")
	assert.NotNil(t, actualNomadAllocs)
	assert.Equal(t, expectedNomadAllocs, actualNomadAllocs)

	httpmock.RegisterResponder("GET", "http://goodAddress/v1/job/job3/allocations",
		httpmock.NewStringResponder(200, `
			[
				invalid JSON
			]`,
		),
	)
	expectedNomadAllocs = nil
	actualNomadAllocs = getNomadAllocs("goodAddress", "job3")
	assert.Empty(t, actualNomadAllocs)
	assert.Equal(t, expectedNomadAllocs, actualNomadAllocs)

	expectedNomadAllocs = nil
	actualNomadAllocs = getNomadAllocs("goodAddress", "badJobID")
	assert.Empty(t, actualNomadAllocs)
	assert.Equal(t, expectedNomadAllocs, actualNomadAllocs)

	expectedNomadAllocs = nil
	actualNomadAllocs = getNomadAllocs("badAddress", "job2")
	assert.Empty(t, actualNomadAllocs)
	assert.Equal(t, expectedNomadAllocs, actualNomadAllocs)
}

func TestGetRSS(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"13459456"
							]
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID/allocations",
		httpmock.NewStringResponder(200, `
			[
				{
					"ID": "alloc_id1"
				},
				{
					"ID": "alloc_id2"
				},
				{
					"ID": "alloc_id3"
				},
				{
					"ID": "alloc_id4"
				}
			]`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=nomad_client_allocs_memory_rss_value",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	expectedRSS := 13459456 / 1.049e6
	expectedRemainders := map[string][]string{
		"alloc_id3": {"rss"},
		"alloc_id4": {"rss"},
	}
	actualRemainders := map[string][]string{}
	actualRSS := getRSS("clusterAddress", "metricsAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualRSS)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedRSS = 13459456 / 1.049e6
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualRSS = getRSS("badAddress", "metricsAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualRSS)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedRSS = 0.0
	expectedRemainders = map[string][]string{
		"alloc_id1": {"rss"},
		"alloc_id2": {"rss"},
		"alloc_id3": {"rss"},
		"alloc_id4": {"rss"},
	}
	actualRemainders = map[string][]string{}
	actualRSS = getRSS("clusterAddress", "badAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualRSS)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedRSS = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualRSS = getRSS("badAddress", "badAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualRSS)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	httpmock.RegisterResponder("GET", "http://metricsAddress2/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				invalid JSON
			}`,
		),
	)
	expectedRSS = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualRSS = getRSS("clusterAddress", "metricsAddress2", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualRSS)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	httpmock.RegisterResponder("GET", "http://metricsAddress3/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"notFloat"
							]
						}
					]
				}
			}`,
		),
	)
	expectedRSS = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualRSS = getRSS("clusterAddress", "metricsAddress3", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualRSS)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)
}

func TestGetCache(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"13459456"
							]
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID/allocations",
		httpmock.NewStringResponder(200, `
			[
				{
					"ID": "alloc_id1"
				},
				{
					"ID": "alloc_id2"
				},
				{
					"ID": "alloc_id3"
				},
				{
					"ID": "alloc_id4"
				}
			]`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=nomad_client_allocs_memory_cache_value",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	expectedCache := 13459456 / 1.049e6
	expectedRemainders := map[string][]string{
		"alloc_id3": {"cache"},
		"alloc_id4": {"cache"},
	}
	actualRemainders := map[string][]string{}
	actualCache := getCache("clusterAddress", "metricsAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedCache, actualCache)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedCache = 13459456 / 1.049e6
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualCache = getCache("badAddress", "metricsAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedCache, actualCache)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedCache = 0.0
	expectedRemainders = map[string][]string{
		"alloc_id1": {"cache"},
		"alloc_id2": {"cache"},
		"alloc_id3": {"cache"},
		"alloc_id4": {"cache"},
	}
	actualRemainders = map[string][]string{}
	actualCache = getCache("clusterAddress", "badAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedCache, actualCache)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedCache = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualCache = getCache("badAddress", "badAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedCache, actualCache)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	httpmock.RegisterResponder("GET", "http://metricsAddress2/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				invalid JSON
			}`,
		),
	)
	expectedCache = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualCache = getCache("clusterAddress", "metricsAddress2", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedCache, actualCache)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	httpmock.RegisterResponder("GET", "http://metricsAddress3/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"notFloat"
							]
						}
					]
				}
			}`,
		),
	)
	expectedCache = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualCache = getCache("clusterAddress", "metricsAddress3", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedCache, actualCache)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)
}

func TestGetTicks(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"13459456"
							]
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID/allocations",
		httpmock.NewStringResponder(200, `
			[
				{
					"ID": "alloc_id1"
				},
				{
					"ID": "alloc_id2"
				},
				{
					"ID": "alloc_id3"
				},
				{
					"ID": "alloc_id4"
				}
			]`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=nomad_client_allocs_cpu_total_ticks_value",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	expectedTicks := 13459456.0
	expectedRemainders := map[string][]string{
		"alloc_id3": {"ticks"},
		"alloc_id4": {"ticks"},
	}
	actualRemainders := map[string][]string{}
	actualTicks := getTicks("clusterAddress", "metricsAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedTicks = 13459456.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualTicks = getTicks("badAddress", "metricsAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedTicks = 0.0
	expectedRemainders = map[string][]string{
		"alloc_id1": {"ticks"},
		"alloc_id2": {"ticks"},
		"alloc_id3": {"ticks"},
		"alloc_id4": {"ticks"},
	}
	actualRemainders = map[string][]string{}
	actualTicks = getTicks("clusterAddress", "badAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	expectedTicks = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualTicks = getTicks("badAddress", "badAddress", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	httpmock.RegisterResponder("GET", "http://metricsAddress2/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				Invalid JSON
			}`,
		),
	)
	expectedTicks = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualTicks = getTicks("clusterAddress", "metricsAddress2", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)

	httpmock.RegisterResponder("GET", "http://metricsAddress3/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"notFloat"
							]
						}
					]
				}
			}`,
		),
	)
	expectedTicks = 0.0
	expectedRemainders = map[string][]string{}
	actualRemainders = map[string][]string{}
	actualTicks = getTicks("clusterAddress", "metricsAddress3", "jobID", "jobName", actualRemainders)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.NotNil(t, actualRemainders)
	assert.Equal(t, expectedRemainders, actualRemainders)
}

func TestGetRemainderNomad(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/client/allocation/alloc_id1/stats",
		httpmock.NewStringResponder(200, `
			{
				"ResourceUsage": {
					"MemoryStats": {
						"RSS": 6451200,
						"Cache": 654321,
						"Swap": 0,
						"Usage": 7569408,
						"MaxUsage": 9162752,
						"KernelUsage": 0,
						"KernelMaxUsage": 0
					},
					"CpuStats": {
						"TotalTicks": 2394.4724337708644
					}
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/client/allocation/alloc_id2/stats",
		httpmock.NewStringResponder(200, `
			{
				"ResourceUsage": {
					"MemoryStats": {
						"RSS": 552821,
						"Cache": 789246,
						"Swap": 0,
						"Usage": 98176514,
						"MaxUsage": 16546,
						"KernelUsage": 0,
						"KernelMaxUsage": 0
					},
					"CpuStats": {
						"TotalTicks": 1125.6842315
					}
				}
			}`,
		),
	)
	remainders := map[string][]string{
		"alloc_id1": {"rss", "cache", "ticks"},
		"alloc_id2": {"rss", "cache", "ticks"},
	}
	expectedRSS := 6451200/1.049e6 + 552821/1.049e6
	expectedCache := 654321/1.049e6 + 789246/1.049e6
	expectedTicks := 2394.4724337708644 + 1125.6842315
	actualRSS, actualCache, actualTicks := getRemainderNomad("clusterAddress", remainders)
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualCache)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)

	remainders = map[string][]string{
		"alloc_id1": {"cache", "ticks"},
		"alloc_id2": {"rss"},
	}
	expectedRSS = 552821 / 1.049e6
	expectedCache = 654321 / 1.049e6
	expectedTicks = 2394.4724337708644
	actualRSS, actualCache, actualTicks = getRemainderNomad("clusterAddress", remainders)
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualCache)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)

	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/client/allocation/alloc_id3/stats",
		httpmock.NewStringResponder(200, `
			{
				Invalid JSON
			}`,
		),
	)
	remainders = map[string][]string{
		"alloc_id3": {"rss", "cache", "ticks"},
	}
	expectedRSS = 0.0
	expectedCache = 0.0
	expectedTicks = 0.0
	actualRSS, actualCache, actualTicks = getRemainderNomad("clusterAddress", remainders)
	assert.Empty(t, actualRSS)
	assert.Empty(t, actualCache)
	assert.Empty(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)

	remainders = map[string][]string{
		"alloc_id1": {"rss", "cache", "ticks"},
		"alloc_id2": {"rss", "cache", "ticks"},
	}
	expectedRSS = 0.0
	expectedCache = 0.0
	expectedTicks = 0.0
	actualRSS, actualCache, actualTicks = getRemainderNomad("badAddress", remainders)
	assert.Empty(t, actualRSS)
	assert.Empty(t, actualCache)
	assert.Empty(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)

	remainders = map[string][]string{
		"alloc_id1": {"rss", "cache", "ticks"},
		"alloc_id3": {"rss", "cache", "ticks"},
	}
	expectedRSS = 6451200 / 1.049e6
	expectedCache = 654321 / 1.049e6
	expectedTicks = 2394.4724337708644
	actualRSS, actualCache, actualTicks = getRemainderNomad("clusterAddress", remainders)
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualCache)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)

	remainders = map[string][]string{
		"alloc_id3": {"rss", "cache", "ticks"},
		"alloc_id1": {"rss", "cache", "ticks"},
	}
	expectedRSS = 6451200 / 1.049e6
	expectedCache = 654321 / 1.049e6
	expectedTicks = 2394.4724337708644
	actualRSS, actualCache, actualTicks = getRemainderNomad("clusterAddress", remainders)
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualCache)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)

	remainders = map[string][]string{
		"alloc_id4": {"rss", "cache", "ticks"},
		"alloc_id1": {"rss", "cache", "ticks"},
	}
	expectedRSS = 6451200 / 1.049e6
	expectedCache = 654321 / 1.049e6
	expectedTicks = 2394.4724337708644
	actualRSS, actualCache, actualTicks = getRemainderNomad("clusterAddress", remainders)
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualCache)
	assert.NotNil(t, actualTicks)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedCache, actualCache)
	assert.Equal(t, expectedTicks, actualTicks)
}

func TestAggUsed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=sum(nomad_client_allocs_memory_rss_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"13459456"
							]
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=sum(nomad_client_allocs_memory_cache_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"33459456"
							]
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=sum(nomad_client_allocs_cpu_total_ticks_value%7Bjob%3D%22jobName%22%7D)%20by%20(job)",
		httpmock.NewStringResponder(200, `
			{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"job": "jobName"
							},
							"value": [
								1597365496,
								"23459456.0"
							]
						}
					]
				}
			}`,
		),
	)
	expectedRSS := 13459456 / 1.049e6
	expectedTicks := 23459456.0
	expectedCache := 33459456 / 1.049e6
	actualRSS, actualTicks, actualCache := aggUsed("clusterAddress", "metricsAddress", "jobID", "jobName")
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualTicks)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.Equal(t, expectedCache, actualCache)

	// nomadAllocs
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID/allocations",
		httpmock.NewStringResponder(200, `
			[
				{
					"ID": "alloc_id1"
				},
				{
					"ID": "alloc_id2"
				}
			]`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/client/allocation/alloc_id1/stats",
		httpmock.NewStringResponder(200, `
			{
				"ResourceUsage": {
					"MemoryStats": {
						"RSS": 6451200,
						"Cache": 654321,
						"Swap": 0,
						"Usage": 7569408,
						"MaxUsage": 9162752,
						"KernelUsage": 0,
						"KernelMaxUsage": 0
					},
					"CpuStats": {
						"TotalTicks": 2394.4724337708644
					}
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/client/allocation/alloc_id2/stats",
		httpmock.NewStringResponder(200, `
			{
				"ResourceUsage": {
					"MemoryStats": {
						"RSS": 552821,
						"Cache": 789246,
						"Swap": 0,
						"Usage": 98176514,
						"MaxUsage": 16546,
						"KernelUsage": 0,
						"KernelMaxUsage": 0
					},
					"CpuStats": {
						"TotalTicks": 1125.6842315
					}
				}
			}`,
		),
	)
	expectedRSS = (13459456 + 6451200 + 552821) / 1.049e6
	expectedTicks = 23459456.0 + 2394.4724337708644 + 1125.6842315
	expectedCache = (33459456 + 654321 + 789246) / 1.049e6
	actualRSS, actualTicks, actualCache = aggUsed("clusterAddress", "metricsAddress", "jobID", "jobName")
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualTicks)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.Equal(t, expectedCache, actualCache)

	// VMAllocs
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=nomad_client_allocs_memory_rss_value",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=nomad_client_allocs_cpu_total_ticks_value",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://metricsAddress/api/v1/query?query=nomad_client_allocs_memory_cache_value",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1"
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2"
							}
						}
					]
				}
			}`,
		),
	)
	expectedRSS = 13459456 / 1.049e6
	expectedTicks = 23459456.0
	expectedCache = 33459456 / 1.049e6
	actualRSS, actualTicks, actualCache = aggUsed("clusterAddress", "metricsAddress", "jobID", "jobName")
	assert.NotNil(t, actualRSS)
	assert.NotNil(t, actualTicks)
	assert.NotNil(t, actualCache)
	assert.Equal(t, expectedRSS, actualRSS)
	assert.Equal(t, expectedTicks, actualTicks)
	assert.Equal(t, expectedCache, actualCache)
}

func TestAggRequested(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// System Job
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID",
		httpmock.NewStringResponder(200, `
			{
				"ID": "jobID",
				"TaskGroups": [
					{
						"Name": "TaskGroup1",
						"Count": 1,
						"Tasks": [
							{
								"Resources": {
									"CPU": 200,
									"MemoryMB": 512,
									"IOPS": 20
								}
							}
						],
						"EphemeralDisk": {
							"SizeMB": 1000
						}
					},
					{
						"Name": "TaskGroup2",
						"Count": 1,
						"Tasks": [
							{
								"Resources": {
									"CPU": 400,
									"MemoryMB": 256,
									"IOPS": 40
								}
							}
						],
						"EphemeralDisk": {
							"SizeMB": 500
						}
					}
				]
			}`,
		),
	)
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID/allocations",
		httpmock.NewStringResponder(200, `
			[
				{
					"TaskGroup": "TaskGroup1"
				},
				{
					"TaskGroup": "TaskGroup1"
				},
				{
					"TaskGroup": "TaskGroup2"
				},
				{
					"TaskGroup": "TaskGroup2"
				},
				{
					"TaskGroup": "TaskGroup2"
				}
			]`,
		),
	)
	expectedCPU := 1600.0
	expectedMemory := 1792.0
	expectedDisk := 3500.0
	expectedIOPS := 160.0
	actualCPU, actualMemory, actualDisk, actualIOPS := aggRequested("clusterAddress", "jobID", "system")
	assert.NotNil(t, actualCPU)
	assert.NotNil(t, actualMemory)
	assert.NotNil(t, actualDisk)
	assert.NotNil(t, actualIOPS)
	assert.Equal(t, expectedCPU, actualCPU)
	assert.Equal(t, expectedMemory, actualMemory)
	assert.Equal(t, expectedDisk, actualDisk)
	assert.Equal(t, expectedIOPS, actualIOPS)

	// Service Job
	httpmock.RegisterResponder("GET", "http://clusterAddress/v1/job/jobID2",
		httpmock.NewStringResponder(200, `
			{
				"ID": "jobID",
				"TaskGroups": [
					{
						"Name": "TaskGroup1",
						"Count": 3,
						"Tasks": [
							{
								"Resources": {
									"CPU": 200,
									"MemoryMB": 512,
									"IOPS": 20
								}
							}
						],
						"EphemeralDisk": {
							"SizeMB": 1000
						}
					},
					{
						"Name": "TaskGroup2",
						"Count": 2,
						"Tasks": [
							{
								"Resources": {
									"CPU": 400,
									"MemoryMB": 256,
									"IOPS": 40
								}
							}
						],
						"EphemeralDisk": {
							"SizeMB": 500
						}
					}
				]
			}`,
		),
	)
	expectedCPU = 1400.0
	expectedMemory = 2048.0
	expectedDisk = 4000.0
	expectedIOPS = 140.0
	actualCPU, actualMemory, actualDisk, actualIOPS = aggRequested("clusterAddress", "jobID2", "service")
	assert.NotNil(t, actualCPU)
	assert.NotNil(t, actualMemory)
	assert.NotNil(t, actualDisk)
	assert.NotNil(t, actualIOPS)
	assert.Equal(t, expectedCPU, actualCPU)
	assert.Equal(t, expectedMemory, actualMemory)
	assert.Equal(t, expectedDisk, actualDisk)
	assert.Equal(t, expectedIOPS, actualIOPS)

	expectedCPU = 0.0
	expectedMemory = 0.0
	expectedDisk = 0.0
	expectedIOPS = 0.0
	actualCPU, actualMemory, actualDisk, actualIOPS = aggRequested("clusterAddress", "jobID", "none")
	assert.NotNil(t, actualCPU)
	assert.NotNil(t, actualMemory)
	assert.NotNil(t, actualDisk)
	assert.NotNil(t, actualIOPS)
	assert.Equal(t, expectedCPU, actualCPU)
	assert.Equal(t, expectedMemory, actualMemory)
	assert.Equal(t, expectedDisk, actualDisk)
	assert.Equal(t, expectedIOPS, actualIOPS)

	expectedCPU = 0.0
	expectedMemory = 0.0
	expectedDisk = 0.0
	expectedIOPS = 0.0
	actualCPU, actualMemory, actualDisk, actualIOPS = aggRequested("badAddress", "jobID", "system")
	assert.NotNil(t, actualCPU)
	assert.NotNil(t, actualMemory)
	assert.NotNil(t, actualDisk)
	assert.NotNil(t, actualIOPS)
	assert.Equal(t, expectedCPU, actualCPU)
	assert.Equal(t, expectedMemory, actualMemory)
	assert.Equal(t, expectedDisk, actualDisk)
	assert.Equal(t, expectedIOPS, actualIOPS)
}