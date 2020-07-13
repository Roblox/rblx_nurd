package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
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

func returnJob(w http.ResponseWriter, r *http.Request) {
	var all []JobDataDB
	vars := mux.Vars(r)
	jobID := vars["id"]
	begin, okBegin := r.URL.Query()["begin"]
	end, okEnd := r.URL.Query()["end"]

	if !okBegin && !okEnd {
		all = getLatestJobDB(db, jobID)
		json.NewEncoder(w).Encode(all)
	} else if !okBegin && okEnd {
		fmt.Fprintf(w, "Missing query param: 'begin'")
	} else if okBegin && !okEnd {
		fmt.Fprintf(w, "Missing query param: 'end'")
	} else {
		all = getTimeSliceDB(db, jobID, begin[0], end[0])
		json.NewEncoder(w).Encode(all)
	}
}

func collectData() {
	addresses, metricsAddress, buffer, duration := loadConfig("config.json")
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

		// printRowsDB(db)
		fmt.Println("done\nElapsed:", end.Sub(begin))
		time.Sleep(duration)
	}
}

func main() {
	go collectData()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/jobs", returnAll)
	router.HandleFunc("/job/{id}", returnJob)
	log.Fatal(http.ListenAndServe(":8080", router))
}
