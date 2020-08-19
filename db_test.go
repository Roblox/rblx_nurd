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
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInitDB(t *testing.T) {
	os.Setenv("CONNECTION_STRING", "VALUE")
	db, insert, err := initDB()
	assert.NotNil(t, err)
	assert.Empty(t, db)

	var DBPtr *sql.DB
	assert.IsType(t, DBPtr, db)
	assert.Empty(t, insert)

	var StmtPtr *sql.Stmt
	assert.IsType(t, StmtPtr, insert)
	assert.Empty(t, StmtPtr)
}

func TestGetAllRowsDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Empty(t, err)
	defer db.Close()

	all, err := getAllRowsDB(nil)
	assert.NotNil(t, err)
	assert.Empty(t, all)

	all, err = getAllRowsDB(db)
	assert.NotNil(t, err)
	assert.Empty(t, all)

	// Test on an empty DB
	query := `SELECT \* FROM resources`
	rows := sqlmock.NewRows([]string{"id", "JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "rIOPS", "namespace", "dataCenters", "date", "insertTime"})
	mock.ExpectQuery(query).WillReturnRows(rows)
	all, err = getAllRowsDB(db)
	assert.Empty(t, err)
	assert.Empty(t, all)

	// Test after inserting rows into DB
	rows = sqlmock.NewRows([]string{"id", "JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "rIOPS", "namespace", "dataCenters", "date", "insertTime"}).
		AddRow(1, "JobID1", "name1", 111.1, 111.1, 111.1, 111.1, 111.1, 111.1, 111.1, "namespace1", "dataCenter1", "0000-00-01", "0000-00-01").
		AddRow(2, "JobID2", "name2", 222.2, 222.2, 222.2, 222.2, 222.2, 222.2, 222.2, "namespace2", "dataCenter2", "0000-00-02", "0000-00-02")
	mock.ExpectQuery(query).WillReturnRows(rows)
	all, err = getAllRowsDB(db)
	assert.Empty(t, err)
	assert.NotEmpty(t, all)

	expected := []JobDataDB{
		JobDataDB{
			"JobID1",
			"name1",
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			"namespace1",
			"dataCenter1",
			"0000-00-01",
			"0000-00-01",
		},
		JobDataDB{
			"JobID2",
			"name2",
			222.2,
			222.2,
			222.2,
			222.2,
			222.2,
			222.2,
			222.2,
			"namespace2",
			"dataCenter2",
			"0000-00-02",
			"0000-00-02",
		},
	}
	assert.Equal(t, expected, all)
}

func TestGetLatestJobDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Empty(t, err)
	defer db.Close()

	all, err := getLatestJobDB(nil, "")
	assert.NotNil(t, err)
	assert.Empty(t, all)

	all, err = getLatestJobDB(db, "")
	assert.NotNil(t, err)
	assert.Empty(t, all)

	// Test on an empty DB
	query := `
		SELECT 
			JobID, 
			name, 
			SUM\(uTicks\), 
			SUM\(rCPU\), 
			SUM\(uRSS\), 
			SUM\(uCache\), 
			SUM\(rMemoryMB\), 
			SUM\(rdiskMB\), 
			namespace, 
			dataCenters, 
			insertTime 
		FROM 
			resources 
		WHERE 
			insertTime IN \(SELECT MAX\(insertTime\) FROM resources\) 
			AND JobID \= 'JobID1' 
		GROUP BY 
			JobID, 
			name, 
			namespace, 
			dataCenters, 
			insertTime`
	rows := sqlmock.NewRows([]string{"JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "namespace", "dataCenters", "insertTime"})
	mock.ExpectQuery(query).WillReturnRows(rows)
	all, err = getLatestJobDB(db, "JobID1")
	assert.Empty(t, err)
	assert.Empty(t, all)

	// Test after inserting rows into DB
	rows = sqlmock.NewRows([]string{"JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "namespace", "dataCenters", "insertTime"}).
		AddRow("JobID1", "name1", 111.1, 111.1, 111.1, 111.1, 111.1, 111.1, "namespace1", "dataCenter1", "0001-01-04T00:00:00Z")
	mock.ExpectQuery(query).WillReturnRows(rows)
	all, err = getLatestJobDB(db, "JobID1")
	assert.Empty(t, err)
	assert.NotEmpty(t, all)

	expected := []JobDataDB{
		JobDataDB{
			"JobID1",
			"name1",
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			0,
			"namespace1",
			"dataCenter1",
			"",
			"0001-01-04T00:00:00Z",
		},
	}
	assert.Equal(t, expected, all)
}

func TestGetTimeSliceDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Empty(t, err)
	defer db.Close()

	all, err := getTimeSliceDB(nil, "", "2020-07-07 17:34:53", "2020-07-18 17:42:19")
	assert.NotNil(t, err)
	assert.Empty(t, all)

	all, err = getTimeSliceDB(db, "", "2020-07-07 17:34:53", "2020-07-18 17:42:19")
	assert.NotNil(t, err)
	assert.Empty(t, all)

	// Test on an empty DB
	rows := sqlmock.NewRows([]string{"JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "namespace", "dataCenters", "insertTime"})
	query := `
		SELECT 
			JobID, 
			name, 
			SUM\(uTicks\), 
			SUM\(rCPU\), 
			SUM\(uRSS\), 
			SUM\(uCache\), 
			SUM\(rMemoryMB\), 
			SUM\(rdiskMB\), 
			namespace, 
			dataCenters, 
			insertTime 
		FROM 
			resources 
		WHERE 
			JobID \= 'JobID1' 
			AND insertTime BETWEEN '2020\-07\-07 17\:34\:53' AND '2020\-07\-18 17\:42\:19' 
		GROUP BY 
			JobID, 
			name, 
			namespace, 
			dataCenters, 
			insertTime 
		ORDER BY 
			insertTime DESC`
	mock.ExpectQuery(query).WillReturnRows(rows)
	all, err = getTimeSliceDB(db, "JobID1", "2020-07-07 17:34:53", "2020-07-18 17:42:19")
	assert.Empty(t, err)
	assert.Empty(t, all)

	// Test after inserting rows into DB
	rows = sqlmock.NewRows([]string{"JobID", "name", "uTicks", "rCPU", "uRSS", "uCache", "rMemoryMB", "rdiskMB", "namespace", "dataCenters", "insertTime"}).
		AddRow("JobID1", "name1", 111.1, 111.1, 111.1, 111.1, 111.1, 111.1, "namespace1", "dataCenter1", "2020-07-07T17:35:00Z")
	mock.ExpectQuery(query).WillReturnRows(rows)
	all, err = getTimeSliceDB(db, "JobID1", "2020-07-07 17:34:53", "2020-07-18 17:42:19")
	assert.Empty(t, err)
	assert.NotEmpty(t, all)

	expected := []JobDataDB{
		JobDataDB{
			"JobID1",
			"name1",
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			111.1,
			0,
			"namespace1",
			"dataCenter1",
			"",
			"2020-07-07T17:35:00Z",
		},
	}
	assert.Equal(t, expected, all)
}
