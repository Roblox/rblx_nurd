package main

import (
	"database/sql"
	"encoding/json"
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

var wg sync.WaitGroup
var db *sql.DB
var insert *sql.Stmt

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to NURD.")
}

func returnAll(w http.ResponseWriter, r *http.Request) {
	logFile, err := os.OpenFile("nurd.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error(err)
	}
	log.SetOutput(logFile)
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)

	all := getAllRowsDB(db, logFile)
	json.NewEncoder(w).Encode(all)

	err = logFile.Close()
	if err != nil {
		log.Error(err)
	}
}

func returnJob(w http.ResponseWriter, r *http.Request) {
	var all []JobDataDB

	logFile, err := os.OpenFile("nurd.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error(err)
	}
	log.SetOutput(logFile)
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)

	jobID := mux.Vars(r)["id"]
	begin, okBegin := r.URL.Query()["begin"]
	end, okEnd := r.URL.Query()["end"]

	if !okBegin && !okEnd {
		all = getLatestJobDB(db, jobID, logFile)
		json.NewEncoder(w).Encode(all)
	} else if !okBegin && okEnd {
		fmt.Fprintf(w, "Missing query param: 'begin'")
	} else if okBegin && !okEnd {
		fmt.Fprintf(w, "Missing query param: 'end'")
	} else {
		all = getTimeSliceDB(db, jobID, begin[0], end[0], logFile)
		json.NewEncoder(w).Encode(all)
	}

	err = logFile.Close()
	if err != nil {
		log.Error(err)
	}
}

func collectData() {
	log.SetReportCaller(true)

	if err := loadConfig("config.json"); err != nil {
		log.Fatal("Error in loading config file")
	}
	db, insert = initDB()
	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Error(err)
	}
	log.SetOutput(logFile)
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)

	db, insert = initDB(logFile)

	for {
		log.Trace("BEGIN AGGREGATION")
		c := make(chan []JobData, len(addresses))

		for _, address := range addresses {
			wg.Add(1)
			go reachCluster(address, metricsAddress, c, logFile)
		}

		wg.Wait()
		close(c)

		// Insert into db from channel
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

	err = logFile.Close()
	if err != nil {
		log.Error(err)
	}
}

func reloadConfig(sigs chan os.Signal) {
	log.SetReportCaller(true)

	for {
		select {
		case <-sigs:
			log.Info("Reloading config file")
			if err := loadConfig("config.json"); err != nil {
				log.Warning("Error in reloading config file")
			}
		default:
		}
	}
}

func main() {
	go collectData()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go reloadConfig(sigs)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/v1/jobs", returnAll)
	router.HandleFunc("/v1/job/{id}", returnJob)
	log.Fatal(http.ListenAndServe(":8080", router))
}
