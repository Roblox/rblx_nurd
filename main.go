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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type APIError struct {
	Error string
}

var (
	wg     sync.WaitGroup
	db     *sql.DB
	insert *sql.Stmt
)

func handleAPIError(w http.ResponseWriter, err string, status int) {
	log.Error(err)
	w.WriteHeader(status)

	errJSON := APIError{
		Error: err,
	}
	if json.NewEncoder(w).Encode(errJSON) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to NURD.")
}

func returnAll(w http.ResponseWriter, r *http.Request) {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)

	all, err := getAllRowsDB(db)
	if err != nil {
		handleAPIError(w, fmt.Sprintf("Error in getting all rows from DB: %v", err), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(all)
	if err != nil {
		handleAPIError(w, fmt.Sprintf("Error in encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func returnJob(w http.ResponseWriter, r *http.Request) {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)

	jobID := mux.Vars(r)["id"]
	begin, okBegin := r.URL.Query()["begin"]
	end, okEnd := r.URL.Query()["end"]

	if !okBegin && !okEnd {
		all, err := getLatestJobDB(db, jobID)
		if err != nil {
			handleAPIError(w, fmt.Sprintf("Error in getting latest job from DB: %v", err), http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(all)
		if err != nil {
			handleAPIError(w, fmt.Sprintf("Error in encoding JSON: %v", err), http.StatusInternalServerError)
			return
		}
	} else if !okBegin && okEnd {
		handleAPIError(w, "Missing query param: 'begin'", http.StatusBadRequest)
	} else if okBegin && !okEnd {
		handleAPIError(w, "Missing query param: 'end'", http.StatusBadRequest)
	} else {
		all, err := getTimeSliceDB(db, jobID, begin[0], end[0])
		if err != nil {
			handleAPIError(w, fmt.Sprintf("Error in getting latest job from DB: %v", err), http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(all)
		if err != nil {
			handleAPIError(w, fmt.Sprintf("Error in encoding JSON: %v", err), http.StatusInternalServerError)
			return;
		}
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)
	w.WriteHeader(http.StatusOK)
}

func collectData(freq *string) {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	duration, err := time.ParseDuration(*freq)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to parse duration: %v", err))
	}
	if duration > 30 * time.Minute || duration <= 0 * time.Minute {
		log.Warning("--aggregate-frequency should be within (0m, 30m]. Defaulting to 15m.")
		duration = 15 * time.Minute
	}

	err = loadConfig("/etc/nurd/config.json")
	if err != nil {
		log.Fatal(fmt.Sprintf("Error in loading /etc/nurd/config.json: %v", err))
	}

	// Retry initializing DB 5 times before exiting
	retryLoad := 5
	for i := 0; i < retryLoad; i++ {
		db, insert, err = initDB()
		if err != nil {
			log.Warning(fmt.Sprintf("DB initialization failed, retrying: %v", err))
		} else {
			log.Info("DB initialized successfully, break ...")
			break
		}

		if i == retryLoad-1 {
			log.Fatal(fmt.Sprintf("Error in initializing DB: %v", err))
		}

		time.Sleep(5 * time.Second)
	}

	for {
		log.Trace("BEGIN AGGREGATION")
		c := make(chan []JobData, len(nomadAddresses))

		for _, address := range nomadAddresses {
			wg.Add(1)
			go reachCluster(address, metricsAddress, c)
		}

		wg.Wait()
		close(c)

		insertTime := time.Now().Truncate(time.Minute).Format("2006-01-02 15:04:05")
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

		log.Trace("END AGGREGATION")
		time.Sleep(duration)
	}
}

func reloadConfig(sigs chan os.Signal) {
	log.SetReportCaller(true)

	for {
		select {
		case <-sigs:
			log.Info("Reloading /etc/nurd/config.json")
			if err := loadConfig("/etc/nurd/config.json"); err != nil {
				log.Warning(fmt.Sprintf("Error in reloading /etc/nurd/config.json: %v", err))
			}
		default:
		}
	}
}

func main() {
	freq := flag.String("aggregate-frequency", "15m", "frequency of resource aggregation")
	flag.Parse()
	go collectData(freq)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go reloadConfig(sigs)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/v1/jobs", returnAll)
	router.HandleFunc("/v1/job/{id}", returnJob)
	router.HandleFunc("/v1/health", healthCheck)
	log.Fatal(http.ListenAndServe(":8080", router))
}
