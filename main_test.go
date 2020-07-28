package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomePage(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	assert.Empty(t, err)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homePage)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
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
