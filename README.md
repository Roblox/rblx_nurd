# <img src="NURD.png" width="30" height="40" alt=":NURD:" class="emoji" title=":NURD:"/> Nomad Usage Resource Dashboard (NURD)
NURD is a dashboard which aggregates and displays CPU and memory resource usage for each job running through specified Hashicorp Nomad servers. The dashboard also displays resources requested by each job, which can be used with resource usage to calculate waste and aid capacity planning. 

## Prerequisites
* At least one active Nomad server
* **Recommended:** A VictoriaMetrics server containing allocation level resource statistics

## Setup
1. **Configuration File**<br>
    a. **nurd/config.json**<br>
        This file contains the configuration information for the Nomad server(s) and the VictoriaMetrics server. Note, any amount of servers can be added to the `Nomad` array.

        {
            "VictoriaMetrics": {
                "URL":      URL for VictoriaMetrics server, 
                "Port":     Port for VictoriaMetrics server
            },
            "Nomad": [
                {
                    "URL":      URL for Nomad server, 
                    "Port":     Port for Nomad server
                }
            ]
        }
2. `$ git clone git@github.com:Roblox/nurd.git`
3. `$ cd nurd`
4. `$ make install`
5. `$ nurd`

## Usage
From `localhost:8080`, or an alternative NURD host address, the user can access several endpoints:

### Home Page
* **`/v1`**<br>
The home page for NURD.
    * **Sample Request**<br>
    `http://localhost:8080/v1`

### List All Jobs
* **`/v1/jobs`**<br>
Lists all job data in NURD.
    * **Sample Request**<br>
    `http://localhost:8080/jobs`

### List Specified Job(s)
* **`/v1/job/:job_id`**<br>
Lists the latest recorded job data for the specified job_id.<br>
**Optional Parameters**<br>
`begin`: Specifies the earliest datetime from which to query.<br>
`end`: Specifies the latest datetime from which to query.<br>
    * **Sample Request**<br>
        * `http://localhost:8080/v1/job/sample_job_id`<br>
        * `http://localhost:8080/v1/job/sample_job_id?begin=2020-07-07%2017:34:53&end=2020-07-08%2017:42:19`
    * **Sample Response**<br>
        ```
        [
            {
                "JobID":"sample-job",
                "Name":"sample-job",
                "UTicks":7318.394561709347,
                "RCPU":1500,
                "URSS":21.542070543374642,
                "UCache":0.4997979027645376,
                "RMemoryMB":768,
                "RdiskMB":900,
                "RIOPS":0,
                "Namespace":"default",
                "DataCenters":"alpha,alpha_test",
                "CurrentTime":"",
                "InsertTime":"2020-07-07T11:49:34Z"
            }
        ]
        ```
### Reload Config File
The user can reload the config file without restarting NURD by sending a SIGHUP signal.<br>
`$ kill -S HUP <PID>`<br>
Once the config file has been reloaded and SIGHUP has been sent to NURD, NURD will complete resource aggregation of the addresses in the previous config file before aggregating on the new addresses. 