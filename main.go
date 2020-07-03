package main

import (
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

func main() {
	addresses, metricsAddress, buffer, duration := Config()

	db, insert := initDB("resources")

	// While loop for scrape frequency
	for {
		c := make(chan []JobData, buffer)
		e := make(chan error)

		// Listen for errors
		go func(e chan error) {
			err := <-e
			log.Fatal("Error: ", err)
		}(e)

		// Goroutines for each cluster address
		for _, address := range addresses {
			wg.Add(1)
			go reachCluster(address, metricsAddress, c, e)
		}

		wg.Wait()
		close(c)

		// Insert into db from channel
		for jobDataSlice := range c {
			for _, v := range jobDataSlice {
				insert.Exec(v.JobID,
					v.name,
					v.uTicks,
					v.pTicks,
					v.rCPU,
					v.uRSS,
					v.pRSS,
					v.rMemoryMB,
					v.rdiskMB,
					v.rIOPS,
					v.namespace,
					v.dataCenters,
					v.currentTime)
			}
		}

		printRowsDB(db)
		time.Sleep(duration)
	}
}
