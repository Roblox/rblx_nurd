package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

// Initialize database
// Configure insert SQL statement
func initDB(nameDB string) (*sql.DB, *sql.Stmt) {
	db, err := sql.Open("sqlite3", nameDB+".out")

	if err != nil {
		log.Fatal("Error:", err)
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
		date DATETIME)`)
	createTable.Exec()

	if err != nil {
		log.Fatal("Error:", err)
	}

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
		date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		log.Fatal("Error:", err)
	}

	return db, insert
}

func printRowsDB(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM resources")

	if err != nil {
		log.Fatal("Error:", err)
	}

	var JobID, name, namespace, dataCenters, currentTime string
	var uTicks, rCPU, uRSS, uCache, rMemoryMB, rdiskMB, rIOPS float64
	var id int

	for rows.Next() {
		rows.Scan(&id, &JobID, &name, &uTicks, &rCPU, &uRSS, &uCache, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currentTime)
		fmt.Println(strconv.Itoa(id)+": ", JobID,
			"\n   ", uTicks,
			"\n   ", rCPU,
			"\n   ", uRSS,
			"\n   ", rMemoryMB)
	}
}
