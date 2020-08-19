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
	req, err := http.NewRequest("POST", "/v1/jobs", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnAll)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	expectedStr := APIError{
		Error: "Error in getting all rows from DB: Nil pointer parameter",
	}
	var actualStr APIError
	err = json.NewDecoder(rr.Body).Decode(&actualStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedStr, actualStr)
}


