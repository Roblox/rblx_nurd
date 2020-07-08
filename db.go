package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
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

func initDB() (*sql.DB, *sql.Stmt) {
	db, err := sql.Open("sqlite3", "resources.out")
	if err != nil {
		log.Fatal(err)
	}

	createTable, err := db.Prepare(`CREATE TABLE IF NOT EXISTS resources (id INTEGER PRIMARY KEY,
		JobID TEXT,
		name TEXT,
		uTicks REAL,
		rCPU REAL, 
		uRSS REAL,
		uCache REAL,
		rMemoryMB REAL,
		rdiskMB REAL,
		rIOPS REAL,
		namespace TEXT,
		dataCenters TEXT,
		date DATETIME,
		insertTime DATETIME)`)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	return db, insert
}

func printRowsDB(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM resources")
	if err != nil {
		log.Fatal(err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64
	var id int

	for rows.Next() {
		rows.Scan(&id, &JobID, &name, &uTicks, &rCPU, &uRSS, &uCache, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currentTime, &insertTime)
		fmt.Println(strconv.Itoa(id)+": ", JobID,
			"\n   ", uTicks,
			"\n   ", rCPU,
			"\n   ", uRSS,
			"\n   ", rMemoryMB)
	}
}

func getAllRowsDB(db *sql.DB) []JobDataDB {
	rows, err := db.Query("SELECT * FROM resources")
	if err != nil {
		log.Fatal(err)
	}

	all := make([]JobDataDB, 0)
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
			insertTime})
	}

	return all
}

func getLatestJobDB(db *sql.DB, jobID string) []JobDataDB {
	jobID = "'" + jobID + "'"
	rows, err := db.Query(`SELECT id, JobID, name, SUM(uTicks), SUM(rCPU), SUM(uRSS), SUM(uCache), SUM(rMemoryMB), SUM(rdiskMB), namespace, dataCenters, insertTime 
						   FROM resources 
						   WHERE insertTime IN (SELECT MAX(insertTime) FROM resources) AND JobID = ` + jobID + ` 
						   GROUP BY JobID`)
	if err != nil {
		log.Fatal(err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64
	var id int

	all := make([]JobDataDB, 0)
	for rows.Next() {
		rows.Scan(&id, &JobID, &name, &uTicks, &rCPU, &uRSS, &uCache, &rMemoryMB, &rdiskMB, &namespace, &dataCenters, &insertTime)
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

	return all
}

func getTimeSliceDB(db *sql.DB, jobID, begin, end string) []JobDataDB {
	jobID = "'" + jobID + "'"
	begin = "'" + begin + "'"
	end = "'" + end + "'"
	rows, err := db.Query(`SELECT JobID, name, SUM(uTicks), SUM(rCPU), SUM(uRSS), SUM(uCache), SUM(rMemoryMB), SUM(rdiskMB), namespace, dataCenters, insertTime 
						   FROM resources 
						   WHERE JobID = ` + jobID + ` AND insertTime BETWEEN ` + begin + ` AND ` + end + ` 
						   GROUP BY JobID, insertTime
						   ORDER BY insertTime DESC`)
	if err != nil {
		log.Fatal(err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64

	all := make([]JobDataDB, 0)
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

	return all
}