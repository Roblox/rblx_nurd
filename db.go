package main

import (
	"database/sql"
	"log"
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

func initDB() (*sql.DB, *sql.Stmt) {
	db, err := sql.Open("mssql", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
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
			insertTime,
		},
		)
	}

	return all
}

func getLatestJobDB(db *sql.DB, jobID string) []JobDataDB {
	jobID = "'" + jobID + "'"
	rows, err := db.Query(`SELECT JobID, name, SUM(uTicks), SUM(rCPU), SUM(uRSS), SUM(uCache), SUM(rMemoryMB), SUM(rdiskMB), namespace, dataCenters, insertTime 
						   FROM resources 
						   WHERE insertTime IN (SELECT MAX(insertTime) FROM resources) AND JobID = ` + jobID + ` 
						   GROUP BY JobID, name, namespace, dataCenters, insertTime`)
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

func getTimeSliceDB(db *sql.DB, jobID, begin, end string) []JobDataDB {
	jobID = "'" + jobID + "'"
	begin = "'" + begin + "'"
	end = "'" + end + "'"
	rows, err := db.Query(`SELECT JobID, name, SUM(uTicks), SUM(rCPU), SUM(uRSS), SUM(uCache), SUM(rMemoryMB), SUM(rdiskMB), namespace, dataCenters, insertTime 
						   FROM resources 
						   WHERE JobID = ` + jobID + ` AND insertTime BETWEEN ` + begin + ` AND ` + end + ` 
						   GROUP BY JobID, name, namespace, dataCenters, insertTime
						   ORDER BY insertTime DESC`)
	if err != nil {
		log.Fatal(err)
	}

	var JobID, name, namespace, dataCenters, currentTime, insertTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64

	all := make([]JobDataDB, 0)
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

	return all
}
