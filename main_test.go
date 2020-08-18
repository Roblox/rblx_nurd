package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"fmt"
)

func TestHomePage(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	testJSON := &Disk {
		SizeMB: 12.0,
	}
	json, _ := json.Marshal(testJSON)
	request, err := http.NewRequest("POST", "/", bytes.NewBuffer(json))


	// request, err := http.NewRequest("GET", "/", nil)
	assert.Empty(t, err)
	request.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homePage)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
	body, _ := ioutil.ReadAll(rr.Body)
	fmt.Println(string(body))

}

func TestReturnAll(t *testing.T) {
	request, err := http.NewRequest("GET", "/v1/jobs", nil)
	assert.Empty(t, err)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnAll)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestReturnJob(t *testing.T) {
	request, err := http.NewRequest("GET", "/v1/job/jobID", nil)
	assert.Empty(t, err)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnJob)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
}
