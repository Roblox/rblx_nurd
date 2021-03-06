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
	"fmt"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

type JobDataDB struct {
	JobID       string
	Name        string
	Ticks       float64
	CPU         float64
	RSS         float64
	Cache       float64
	MemoryMB    float64
	diskMB      float64
	IOPS        float64
	Namespace   string
	DataCenters string
	CurrentTime string
	InsertTime  string
}

func initDB() (*sql.DB, *sql.Stmt, error) {
	db, err := sql.Open("mssql", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		return nil, nil, fmt.Errorf("Error in opening DB: %v", err)
	}

	createTable, err := db.Prepare(`if not exists (select * from sysobjects where name='resources' and xtype='U')
		CREATE TABLE resources 
		(id INTEGER IDENTITY(1,1) PRIMARY KEY,
		JobID VARCHAR(255),
		name VARCHAR(255),
		uTicks REAL,
		rCPU REAL, 
		uRSS REAL,
		uCache REAL,
		rMemoryMB REAL,
		rdiskMB REAL,
		rIOPS REAL,
		namespace VARCHAR(255),
		dataCenters VARCHAR(255),
		date DATETIME,
		insertTime DATETIME);`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error in creating DB table: %v", err)
	}
	createTable.Exec()

	insert, err := db.Prepare(`INSERT INTO resources (JobID,
		name,
		uTicks,
		rCPU,
		uRSS,
		uCache,
		rMemoryMB,
		rdiskMB,
		rIOPS,
		namespace,
		dataCenters,
		date,
		insertTime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error in preparing DB insert: %v", err)
	}

	return db, insert, nil
}

func getAllRowsDB(db *sql.DB) ([]JobDataDB, error) {
	if db == nil {
		return nil, fmt.Errorf("Parameter db *sql.DB is nil")
	}

	all := make([]JobDataDB, 0)

	rows, err := db.Query("SELECT * FROM resources")
	if err != nil {
		return nil, fmt.Errorf("Error in querying DB: %v", err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64
	var id int
	for rows.Next() {
		rows.Scan(&id, &JobID, &name, &uTicks, &rCPU, &uRSS, &uCache, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currentTime, &insertTime)
		all = append(all, JobDataDB{
			JobID,
			name,
			uTicks,
			rCPU,
			uRSS,
			uCache,
			rMemoryMB,
			rdiskMB,
			rIOPS,
			namespace,
			dataCenters,
			currentTime,
			insertTime,
		},
		)
	}

	return all, nil
}

func getLatestJobDB(db *sql.DB, jobID string) ([]JobDataDB, error) {
	if db == nil {
		return nil, fmt.Errorf("Parameter db *sql.DB is nil")
	}

	all := make([]JobDataDB, 0)

	jobID = "'" + jobID + "'"
	rows, err := db.Query(`SELECT JobID, name, SUM(uTicks), SUM(rCPU), SUM(uRSS), SUM(uCache), SUM(rMemoryMB), SUM(rdiskMB), namespace, dataCenters, insertTime 
						   FROM resources 
						   WHERE insertTime IN (SELECT MAX(insertTime) FROM resources) AND JobID = ` + jobID + ` 
						   GROUP BY JobID, name, namespace, dataCenters, insertTime`)
	if err != nil {
		return nil, fmt.Errorf("Error in querying DB: %v", err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64

	for rows.Next() {
		rows.Scan(&JobID, &name, &uTicks, &rCPU, &uRSS, &uCache, &rMemoryMB, &rdiskMB, &namespace, &dataCenters, &insertTime)
		all = append(all, JobDataDB{
			JobID,
			name,
			uTicks,
			rCPU,
			uRSS,
			uCache,
			rMemoryMB,
			rdiskMB,
			rIOPS,
			namespace,
			dataCenters,
			currentTime,
			insertTime})
	}

	return all, nil
}

func getTimeSliceDB(db *sql.DB, jobID, begin, end string) ([]JobDataDB, error) {
	if db == nil {
		return nil, fmt.Errorf("Parameter db *sql.DB is nil")
	}

	all := make([]JobDataDB, 0)

	jobID = "'" + jobID + "'"
	begin = "'" + begin + "'"
	end = "'" + end + "'"
	rows, err := db.Query(`SELECT JobID, name, SUM(uTicks), SUM(rCPU), SUM(uRSS), SUM(uCache), SUM(rMemoryMB), SUM(rdiskMB), namespace, dataCenters, insertTime 
						   FROM resources 
						   WHERE JobID = ` + jobID + ` AND insertTime BETWEEN ` + begin + ` AND ` + end + ` 
						   GROUP BY JobID, name, namespace, dataCenters, insertTime
						   ORDER BY insertTime DESC`)
	if err != nil {
		return nil, fmt.Errorf("Error in querying DB: %v", err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64

	for rows.Next() {
		rows.Scan(&JobID, &name, &uTicks, &rCPU, &uRSS, &uCache, &rMemoryMB, &rdiskMB, &namespace, &dataCenters, &insertTime)
		all = append(all,
			JobDataDB{
				JobID,
				name,
				uTicks,
				rCPU,
				uRSS,
				uCache,
				rMemoryMB,
				rdiskMB,
				rIOPS,
				namespace,
				dataCenters,
				currentTime,
				insertTime,
			},
		)
	}

	return all, nil
}
