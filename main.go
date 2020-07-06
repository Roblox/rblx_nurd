package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"fmt"
)

var wg sync.WaitGroup
var db *sql.DB
var insert *sql.Stmt

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to NURD.")
}

func returnAll(w http.ResponseWriter, r *http.Request) {
	all := getAllRowsDB(db)
	json.NewEncoder(w).Encode(all)
}

func handleRequests() {
	http.HandleFunc("/nurd", homePage)
	http.HandleFunc("/nurd/all", returnAll)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func collectData() {
	addresses, metricsAddress, buffer, duration := Config("config.json")
	db, insert = initDB()

	// While loop for scrape frequency
	for {
		c := make(chan []JobData, buffer)
		e := make(chan error)

		// Listen for errors
		go func(e chan error) {
			err := <-e
			log.Fatal(err)
		}(e)

		begin := time.Now()

		// Goroutines for each cluster address
		for _, address := range addresses {
			wg.Add(1)
			go reachCluster(address, metricsAddress, c, e)
		}

		wg.Wait()
		close(c)

		end := time.Now()

		// Insert into db from channel
		insertTime := time.Now().Format("2006-01-02 15:04:05")
		for jobDataSlice := range c {
			for _, v := range jobDataSlice {
				insert.Exec(v.JobID,
					v.Name,
					v.UTicks,
					v.RCPU,
					v.URSS,
					v.UCache,
					v.RMemoryMB,
					v.RdiskMB,
					v.RIOPS,
					v.Namespace,
					v.DataCenters,
					v.CurrentTime,
					insertTime)
			}
		}

		printRowsDB(db)
		fmt.Println("done\nElapsed:", end.Sub(begin))
		time.Sleep(duration)
	}
}

func main() {
	go collectData()
	handleRequests()
}
