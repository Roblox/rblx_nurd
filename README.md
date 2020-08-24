# <img src="NURD.png" width="60" height="80" alt=":NURD:" class="emoji" title=":NURD:"/> Nomad Usage Resource Dashboard (NURD)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/Roblox/rblx_nurd/blob/master/LICENSE)
[![CircleCI](https://circleci-github.rcs.simulpong.com/gh/Roblox/rblx_nurd/tree/master.svg?style=shield&circle-token=638e19f15c88268832a4f2a7bfee4f081df8d65d)](https://circleci-github.rcs.simulpong.com/gh/Roblox/rblx_nurd/tree/master)

NURD is a dashboard which aggregates and displays CPU and memory resource usage for each job running through specified Hashicorp Nomad servers. The dashboard also displays resources requested by each job, which can be used with resource usage to calculate waste and aid capacity planning. 

## Prerequisites
* Docker Version: >=19.03.8+
* **Required:** At least one active Nomad server
* **Optional:** A VictoriaMetrics server containing allocation level resource statistics

## Setup
The user can configure NURD to connect to a containerized SQL Server instance with [docker-compose.yml](https://github.com/Roblox/rblx_nurd/blob/master/docker-compose.yml) or point to another SQL Server instance with [Dockerfile](https://github.com/Roblox/rblx_nurd/blob/master/Dockerfile). See options below for details. By default, NURD collects data every 15 minutes. To modify the frequency, edit [Dockerfile](https://github.com/Roblox/rblx_nurd/blob/master/Dockerfile#L21) with the following formatting style before startup:<br>
`CMD ["nurd", "--aggregate-frequency", "15m"]`

### Containerized SQL Server Instance
1. `$ git clone git@github.com:Roblox/nurd.git`
2. **Configuration**<br>
    * **[docker-compose.yml](https://github.com/Roblox/rblx_nurd/blob/master/docker-compose.yml)**<br>
        This file contains the necessary login information to create a SQL Server instance.
    * **[etc/nurd/config.json](https://github.com/Roblox/rblx_nurd/blob/master/etc/nurd/config.json)**<br>
        This file contains the configuration information for the Nomad server(s) and the VictoriaMetrics server. The default URLs and ports must be overwritten. If no VictoriaMetrics server exists, the VictoriaMetrics stanza must be removed. Note, any amount of servers can be added to the `Nomad` array.
4. `$ docker-compose build`
5. `$ docker-compose up -d`
6. **Grafana Dashboard**<br>
    a. Navigate to [localhost:3000](http://localhost:3000)<br>
    b. Login with
        
        username: admin
        password: admin
    c. Change the password<br>
    d. Navigate to [localhost:3000/datasources/new](http://localhost:3000/datasources/new) and select `Microsoft SQL Server`<br>
    e. Input the following connection data

        Host: mssql
        Database: master
        User: sa
        Password: yourStrong(!)Password
    f. Select `Save & Test`<br>
    g. Navigate to [localhost:3000/dashboard/import](http://localhost:3000/dashboard/import) and select `Upload JSON file`<br>
    h. Upload [grafana.json](https://github.com/Roblox/rblx_nurd/blob/master/grafana.json) and select `import`<br>


### Another SQL Server Instance
1. `$ git clone git@github.com:Roblox/nurd.git`
2. **Configuration**<br>
    * **[Dockerfile](https://github.com/Roblox/rblx_nurd/blob/master/Dockerfile)**<br>
        This file contains the necessary login information to connect to a separate SQL Server instance. It is necessary to configure the [connection string](https://github.com/Roblox/rblx_nurd/blob/master/Dockerfile#L5)  environment variable.
    * **[etc/nurd/config.json](https://github.com/Roblox/rblx_nurd/blob/master/etc/nurd/config.json)**<br>
        This file contains the configuration information for the Nomad server(s) and the VictoriaMetrics server. The default URLs and ports must be overwritten. If no VictoriaMetrics server exists, the VictoriaMetrics stanza must be removed. Note, any amount of servers can be added to the `Nomad` array.
3. `$ cd nurd`
4. `$ docker build -t nurd .`
5. `$ docker run -dp 8080:8080 nurd`

## Exit
1. `$ docker-compose down` __or__ `$ docker stop`

## Usage
### Grafana Dashboard
From [localhost:3000](http://localhost:3000), or an alternative NURD host address, the user can access the Grafana dashboard. The following parameters are available to query through the dropdown menu.<br>
**Note:** No time series will display until NURD has inserted data into the database.<br>
* `JobID`: ID of a job
* `Metrics`
    * `UsedMemory`: the memory currently in use by the selected jobs in MiB
    * `RequestedMemory`: the memory requested by the selected jobs in MiB
    * `UsedCPU`: the CPU currently in use by the selected jobs in MHz
    * `RequestedCPU`: the CPU requested by the selected jobs in MHz
* `Total`: toggle to aggregate metrics over the current selection

### API
From [localhost:8080](http://localhost:8080), or an alternative NURD host address, the user can access several endpoints:

#### Home Page
* **`/`**<br>
The home page for NURD.
    * **Sample Request**<br>
    `http://localhost:8080/`

#### List All Jobs
* **`/v1/jobs`**<br>
Lists all job data in NURD.
    * **Sample Request**<br>
    `http://localhost:8080/v1/jobs`

#### List Specified Job(s)
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
                "DataCenters":"DC0,DC1",
                "CurrentTime":"",
                "InsertTime":"2020-07-07T11:49:34Z"
            }
        ]
        ```
### Reload Config File
NURD supports hot reloading to point NURD to different Nomad clusters and/or a VictoriaMetrics server.

1. `Exec` into the container running NURD<br>
    `$ docker exec -it nurd /bin/bash`
2. Edit the contents of [/etc/nurd/config.json](https://github.com/Roblox/rblx_nurd/blob/master/etc/nurd/config.json)<br>
    `$ vim /etc/nurd/config.json`
3. Exit the container<br>
    `$ exit`
3. Send a SIGHUP signal to the container running NURD.<br>
    `$ docker kill --signal=HUP nurd`

Once SIGHUP has been sent to NURD, NURD will complete resource aggregation of the addresses in the previous cycle before aggregating on the new addresses. 
