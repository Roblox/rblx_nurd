package main

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"fmt"
)

func TestGetPromAllocs(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "http://prom/api/v1/query?query=sampleQuery",
		httpmock.NewStringResponder(200, `
			{
				"status":"successTest",
				"data":{
					"resultType":"vectorTest",
					"result":[
						{
							"metric":{
								"alloc_id":"alloc_id1",
								"value":[0,"0"]
							}
						},
						{
							"metric":{
								"alloc_id":"alloc_id2",
								"value":[0,"0"]
							}
						}
					]
				}
			}`
		)
	)

	m := getPromAllocs("prom", "sampleQuery")
	expected := map[string]struct{}{
		"alloc_id1":{},
		"alloc_id2":{},
	}
	assert.Equal(t, expected, m)
}

