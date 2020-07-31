package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	// "fmt"
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
	var mock sqlmock.Sqlmock
	var err error
	db, mock, err = sqlmock.New()
	assert.Empty(t, err)
	defer db.Close()

	// Empty DB
	rows := sqlmock.NewRows([]string{"id", "JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "rIOPS", "namespace", "dataCenters", "date", "insertTime"})
	mock.ExpectQuery(`SELECT \* FROM resources`).WillReturnRows(rows)

	request, err := http.NewRequest("GET", "/v1/jobs", nil)
	assert.Empty(t, err)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnAll)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)

	// fmt.Println("Result:", rr.Body)

	// Non-empty in DB
	rows = sqlmock.NewRows([]string{"id", "JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "rIOPS", "namespace", "dataCenters", "date", "insertTime"}).
		AddRow(1, "JobID1", "name1", 111.1, 111.1, 111.1, 111.1, 111.1, 111.1, 111.1, "namespace1", "dataCenter1", "0000-00-01", "0000-00-01").
		AddRow(2, "JobID2", "name2", 222.2, 222.2, 222.2, 222.2, 222.2, 222.2, 222.2, "namespace2", "dataCenter2", "0000-00-02", "0000-00-02")
	mock.ExpectQuery(`SELECT \* FROM resources`).WillReturnRows(rows)

	request, err = http.NewRequest("GET", "/v1/jobs", nil)
	assert.Empty(t, err)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(returnAll)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)

	// fmt.Println("Result:", rr.Body)
}

func TestReturnJob(t *testing.T) {
	request, err := http.NewRequest("GET", "/v1/job/jobID", nil)
	assert.Empty(t, err)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnJob)
	handler.ServeHTTP(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code)
}
