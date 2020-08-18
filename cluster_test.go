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
		"alloc_id3": []string{"rss"},
		"alloc_id4": []string{"rss"},
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
		"alloc_id1": []string{"rss"},
		"alloc_id2": []string{"rss"},
		"alloc_id3": []string{"rss"},
		"alloc_id4": []string{"rss"},
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
		"alloc_id3": []string{"cache"},
		"alloc_id4": []string{"cache"},
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
		"alloc_id1": []string{"cache"},
		"alloc_id2": []string{"cache"},
		"alloc_id3": []string{"cache"},
		"alloc_id4": []string{"cache"},
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
		"alloc_id3": []string{"ticks"},
		"alloc_id4": []string{"ticks"},
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
		"alloc_id1": []string{"ticks"},
		"alloc_id2": []string{"ticks"},
		"alloc_id3": []string{"ticks"},
		"alloc_id4": []string{"ticks"},
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
