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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHomePage(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homePage)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	expectedStr := "Welcome to NURD."
	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, string(body))
}

func TestReturnAllNoDB(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/jobs", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnAll)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	expectedStr := APIError{
		Error: "Error in getting all rows from DB: Parameter db *sql.DB is nil",
	}
	var actualStr APIError
	err = json.NewDecoder(rr.Body).Decode(&actualStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, actualStr)
}

func TestReturnJobNoParam(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/job/jobID", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnJob)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	expectedStr := APIError{
		Error: "Error in getting latest job from DB: Parameter db *sql.DB is nil",
	}
	var actualStr APIError
	err = json.NewDecoder(rr.Body).Decode(&actualStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, actualStr)
}

func TestReturnJobNoBegin(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/job/jobID?end=2020-07-18%2017:42:19", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnJob)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	expectedStr := APIError{
		Error: "Missing query param: 'begin'",
	}
	var actualStr APIError
	err = json.NewDecoder(rr.Body).Decode(&actualStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, actualStr)
}

func TestReturnJobNoEnd(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/job/jobID?begin=2020-07-18%2017:42:19", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnJob)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	expectedStr := APIError{
		Error: "Missing query param: 'end'",
	}
	var actualStr APIError
	err = json.NewDecoder(rr.Body).Decode(&actualStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, actualStr)
}

func TestReturnJobParams(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/job/jobID?begin=2020-07-18%2017:42:19&end=2020-07-18%2017:42:20", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnJob)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	expectedStr := APIError{
		Error: "Error in getting latest job from DB: Parameter db *sql.DB is nil",
	}
	var actualStr APIError
	err = json.NewDecoder(rr.Body).Decode(&actualStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, actualStr)
}