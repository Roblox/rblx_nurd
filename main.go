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

var (
	wg     sync.WaitGroup
	db     *sql.DB
	insert *sql.Stmt
)

func homePage(w http.ResponseWriter, r *http.Request) {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)
	
	fmt.Fprintf(w, "Welcome to NURD.")
}

func returnAll(w http.ResponseWriter, r *http.Request) {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	log.Trace(r)

	all, err := getAllRowsDB(db)
	if err != nil {
		log.Error(fmt.Sprintf("Error in getting all rows from DB: %v", err))
	}
	err = json.NewEncoder(w).Encode(all)
	if err != nil {
		log.Error(fmt.Sprintf("Error in encoding JSON: %v", err))
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
			log.Error(fmt.Sprintf("Error in getting latest job from DB: %v", err))
		}
		err = json.NewEncoder(w).Encode(all)
		if err != nil {
			log.Error(fmt.Sprintf("Error in encoding JSON: %v", err))
		}
	} else if !okBegin && okEnd {
		fmt.Fprintf(w, "Missing query param: 'begin'")
	} else if okBegin && !okEnd {
		fmt.Fprintf(w, "Missing query param: 'end'")
	} else {
		all, err := getTimeSliceDB(db, jobID, begin[0], end[0])
		if err != nil {
			log.Error(fmt.Sprintf("Error in getting latest job from DB: %v", err))
		}
		err = json.NewEncoder(w).Encode(all)
		if err != nil {
			log.Error(fmt.Sprintf("Error in encoding JSON: %v", err))
		}
	}
}

func collectData() {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	err := loadConfig("/etc/nurd/config.json")
	if err != nil {
		log.Fatal(fmt.Sprintf("Error in loading /etc/nurd/config.json: %v", err))
	}

	for i := 0; i < 10; i++ {
		db, insert, err = initDB()
		if err == nil {
			break
		}

		if i == 9 {
			log.Fatal(fmt.Sprintf("Error in initializing DB: %v", err))
		}

		log.Warning(fmt.Sprintf("Waiting to initialize DB: %v", err))
		duration, err := time.ParseDuration("5s")
		if err != nil {
			log.Error(fmt.Sprintf("Error in parsing duration: %v", err))
		}
		time.Sleep(duration)
	}

	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Error(fmt.Sprintf("Error in parsing duration: %v", err))
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
			log.Info("Reloading config file")
			if err := loadConfig("/etc/nurd/config.json"); err != nil {
				log.Warning(fmt.Sprintf("Error in reloading /etc/nurd/config.json: %v", err))
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
