package main

import (
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"fmt"
)

var wg sync.WaitGroup

func main() {
	addresses, metricsAddress, buffer, duration := Config("config.txt")

	db, insert := initDB()

	// While loop for scrape frequency
	for {
		c := make(chan []JobData, buffer)
		e := make(chan error)

		// Listen for errors
		go func(e chan error) {
			err := <-e
			log.Fatal(err)
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
					v.rCPU,
					v.uRSS,
					v.uCache,
					v.rMemoryMB,
					v.rdiskMB,
					v.rIOPS,
					v.namespace,
					v.dataCenters,
					v.currentTime)
			}
		}

		printRowsDB(db)
		fmt.Println("done")
		time.Sleep(duration)
	}
}
