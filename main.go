package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

func main() {
	addresses, buffer, duration := Config()

	db, insert := initDB("resources")

	// While loop for scrape frequency
	for {
		c := make(chan []JobData, buffer)
		e := make(chan error)
		// m := make(map[string]JobData)

		begin := time.Now()

		// Listen for errors
		go func(e chan error) {
			err := <-e
			log.Fatal("Error: ", err)
		}(e)

		// Goroutines for each cluster address
		for _, address := range addresses {
			wg.Add(1)
			go reachCluster(address, c, e)
		}

		wg.Wait()
		close(c)

		end := time.Now()

		// Insert into db from channel
		for jobDataSlice := range c {
			for _, v := range jobDataSlice {
				insert.Exec(v.JobID,
					v.uTicks,
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
		fmt.Println("Elapsed:", end.Sub(begin))
		time.Sleep(duration)
	}
}
