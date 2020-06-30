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
	db, errOpen := sql.Open("sqlite3", nameDB+".out")

	if errOpen != nil {
		log.Fatal("Error:", errOpen)
	}

	createTable, errCreateTable := db.Prepare(`CREATE TABLE IF NOT EXISTS resources (id INTEGER PRIMARY KEY,
		JobID TEXT,
		uTicks REAL,
		rCPU REAL, 
		uRSS REAL,
		pRSS REAL,
		rMemoryMB REAL,
		rdiskMB REAL,
		rIOPS REAL,
		namespace TEXT,
		dataCenters TEXT,
		date DATETIME)`)
	createTable.Exec()

	if errCreateTable != nil {
		log.Fatal("Error:", errCreateTable)
	}

	insert, errInsert := db.Prepare(`INSERT INTO resources (JobID,
		uTicks, 
		rCPU,
		uRSS,
		pRSS,
		rMemoryMB,
		rdiskMB,
		rIOPS,
		namespace,
		dataCenters,
		date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if errInsert != nil {
		log.Fatal("Error:", errInsert)
	}

	return db, insert
}

func printRowsDB(db *sql.DB) {
	rows, _ := db.Query("SELECT * FROM resources")

	var JobID, namespace, dataCenters, currentTime string
	var uTicks, rCPU, uRSS, rMemoryMB, rdiskMB, rIOPS, pRSS float64
	var id int

	for rows.Next() {
		rows.Scan(&id, &JobID, &uTicks, &rCPU, &uRSS, &pRSS, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currentTime)
		fmt.Println(strconv.Itoa(id)+": ", JobID, uTicks, rCPU, uRSS, pRSS, rMemoryMB, rdiskMB, rIOPS, namespace, dataCenters, currentTime)
	}
}
