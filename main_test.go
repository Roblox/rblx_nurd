package main

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	add, buf, dur := config()

	if reflect.TypeOf(add[0]).Kind() != reflect.String {
		t.Errorf("\nExpected: string\nActual: %T", add[0])
	}

	if buf != 1 {
		t.Errorf("\nExpected: 1\nActual: %d", buf)
	}

	minute, _ := time.ParseDuration("1m")
	if dur != minute {
		t.Errorf("\nExpected: 60000000000\nActual: %d", dur)
	}
}

func TestInitDB(t *testing.T) {
	db, insert := initDB("test")

	_, errOpen := os.Open("test.db")
	if errOpen != nil {
		t.Errorf("\nExpected: nil\nActual: %d", errOpen)
	}

	errPing := db.Ping()
	if errPing != nil {
		t.Errorf("\nExpected: nil\nActual: %d", errPing)
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	result, errInsert := insert.Exec("JobID", 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, "namespace", "DC", currentTime)
	if errInsert != nil {
		t.Errorf("\nExpected: nil\nActual: %d", errInsert)
	}

	num, _ := result.RowsAffected()
	if num != 1 {
		t.Errorf("\nExpected: 1\nActual: %d", num)
	}

	rows, _ := db.Query("SELECT * FROM resources")
	columnsActual, _ := rows.Columns()
	columnsExpected := []string{"id", "JobID", "uTicks", "rCPU", "uRSS", "rMemoryMB", "rdiskMB", "rIOPS", "namespace", "dataCenters", "date"}
	if len(columnsActual) != 11 {
		t.Errorf("\nExpected: 11\nActual: %d", len(columnsActual))
	}

	for i := range columnsExpected {
		if columnsActual[i] != columnsExpected[i] {
			t.Errorf("\nExpected: %v\nActual: %v", columnsExpected[i], columnsActual[i])
		}
	}

	var JobID, namespace, dataCenters, currTime string
	var uTicks, rCPU, uRSS, rMemoryMB, rdiskMB, rIOPS float64
	var id int

	for rows.Next() {
		rows.Scan(&id, &JobID, &uTicks, &rCPU, &uRSS, &rMemoryMB, &rdiskMB, &rIOPS, &namespace, &dataCenters, &currTime)
	}
	nowTime := time.Now().Format("2006-01-02T15:04:05Z")
	if currTime != nowTime {
		t.Errorf("\nExpected: %v\nActual: %v", currTime, nowTime)
	}
}