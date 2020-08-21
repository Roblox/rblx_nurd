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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func populateDB(t *testing.T, insert *sql.Stmt) {
	layout := "2006-01-02 15:04:05"
	str := "2000-01-01 00:00:00"
	time1, err := time.Parse(layout, str)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insert.Exec("JobID1", "JobName1", 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, "Namespace1", "DC1", time1, time1)
	if err != nil {
		t.Fatal(err)
	}

	str = "2000-01-02 00:00:00"
	time2, err := time.Parse(layout, str)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insert.Exec("JobID1", "JobName1", 3.0, 3.0, 3.0, 3.0, 3.0, 3.0, 3.0, "Namespace1", "DC1", time2, time2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insert.Exec("JobID1", "JobName1", 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, "Namespace1", "DC1", time2, time2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = insert.Exec("JobID2", "JobName2", 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, "Namespace2", "DC2", time2, time2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitDBMock(t *testing.T) {
	var DBPtr *sql.DB
	var StmtPtr *sql.Stmt

	os.Setenv("CONNECTION_STRING", "VALUE")

	db, insert, err := initDB()
	assert.NotNil(t, err)
	assert.Empty(t, db)
	assert.IsType(t, DBPtr, db)
	assert.Empty(t, insert)
	assert.IsType(t, StmtPtr, insert)
	assert.Empty(t, insert)
}

func TestInitDBLive(t *testing.T) {
	var db, DBPtr *sql.DB
	var insert, StmtPtr *sql.Stmt
	var err error

	os.Setenv("CONNECTION_STRING", "Server=localhost;Database=master;User Id=sa;Password=yourStrong(!)Password;")

	// Retry initializing DB 5 times before failing
	retryLoad := 5
	for i := 0; i < retryLoad; i++ {
		db, insert, err = initDB()
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	assert.Empty(t, err)
	assert.NotNil(t, db)
	assert.IsType(t, DBPtr, db)
	assert.NotNil(t, insert)
	assert.IsType(t, StmtPtr, insert)
	assert.NotNil(t, insert)
}

func TestGetAllRowsDBMock(t *testing.T) {
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
		{
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
		{
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

func TestGetAllRowsDBLive(t *testing.T) {
	var db *sql.DB
	var insert *sql.Stmt
	var err error

	os.Setenv("CONNECTION_STRING", "Server=localhost;Database=master;User Id=sa;Password=yourStrong(!)Password;")
	db, insert, err = initDB()

	populateDB(t, insert)

	all, err := getAllRowsDB(db)
	assert.Nil(t, err)
	assert.NotNil(t, all)

	expected := []JobDataDB{
		{
			"JobID1",
			"JobName1",
			1.0,
			1.0,
			1.0,
			1.0,
			1.0,
			1.0,
			1.0,
			"Namespace1",
			"DC1",
			"2000-01-01T00:00:00Z",
			"2000-01-01T00:00:00Z",
		},
		{
			"JobID1",
			"JobName1",
			3.0,
			3.0,
			3.0,
			3.0,
			3.0,
			3.0,
			3.0,
			"Namespace1",
			"DC1",
			"2000-01-02T00:00:00Z",
			"2000-01-02T00:00:00Z",
		},
		{
			"JobID1",
			"JobName1",
			2.0,
			2.0,
			2.0,
			2.0,
			2.0,
			2.0,
			2.0,
			"Namespace1",
			"DC1",
			"2000-01-02T00:00:00Z",
			"2000-01-02T00:00:00Z",
		},
		{
			"JobID2",
			"JobName2",
			2.0,
			2.0,
			2.0,
			2.0,
			2.0,
			2.0,
			2.0,
			"Namespace2",
			"DC2",
			"2000-01-02T00:00:00Z",
			"2000-01-02T00:00:00Z",
		},
	}
	assert.Equal(t, expected, all)
}

func TestGetLatestJobDBMock(t *testing.T) {
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
		{
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

func TestGetLatestJobDBLive(t *testing.T) {
	var db *sql.DB
	var err error

	os.Setenv("CONNECTION_STRING", "Server=localhost;Database=master;User Id=sa;Password=yourStrong(!)Password;")
	db, _, err = initDB()

	all, err := getLatestJobDB(db, "JobID1")
	assert.Nil(t, err)
	assert.NotNil(t, all)
	expected := []JobDataDB{
		{
			"JobID1",
			"JobName1",
			5.0,
			5.0,
			5.0,
			5.0,
			5.0,
			5.0,
			0.0,
			"Namespace1",
			"DC1",
			"",
			"2000-01-02T00:00:00Z",
		},
	}
	assert.Equal(t, expected, all)
}

func TestGetTimeSliceDBMock(t *testing.T) {
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
		{
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

func TestGetTimeSliceDBLive(t *testing.T) {
	var db *sql.DB
	var err error

	os.Setenv("CONNECTION_STRING", "Server=localhost;Database=master;User Id=sa;Password=yourStrong(!)Password;")
	db, _, err = initDB()

	all, err := getTimeSliceDB(db, "JobID1", "2000-01-01 00:00:01", "2000-01-02 00:00:01")
	assert.Nil(t, err)
	assert.NotNil(t, all)

	expected := []JobDataDB{
		{
			"JobID1",
			"JobName1",
			5.0,
			5.0,
			5.0,
			5.0,
			5.0,
			5.0,
			0,
			"Namespace1",
			"DC1",
			"",
			"2000-01-02T00:00:00Z",
		},
	}
	assert.Equal(t, expected, all)

	all, err = getTimeSliceDB(db, "JobID1", "2000-01-01 00:00:00", "2000-01-01 12:00:01")
	assert.Nil(t, err)
	assert.NotNil(t, all)

	expected = []JobDataDB{
		{
			"JobID1",
			"JobName1",
			1.0,
			1.0,
			1.0,
			1.0,
			1.0,
			1.0,
			0,
			"Namespace1",
			"DC1",
			"",
			"2000-01-01T00:00:00Z",
		},
	}
	assert.NotNil(t, all)
	assert.Equal(t, expected, all)

	all, err = getTimeSliceDB(db, "JobID1", "2000-04-04 00:00:00", "2000-05-05 12:00:01")
	assert.Nil(t, err)
	assert.NotNil(t, all)

	expected = []JobDataDB{}
	assert.NotNil(t, all)
	assert.Equal(t, expected, all)
}