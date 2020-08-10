# <img src="NURD.png" width="33" height="44" alt=":NURD:" class="emoji" title=":NURD:"/> Nomad Usage Resource Dashboard (NURD)
NURD is a dashboard which aggregates and displays CPU and memory resource usage for each job running through specified Hashicorp Nomad servers. The dashboard also displays resources requested by each job, which can be used with resource usage to calculate waste and aid capacity planning. 

## Prerequisites
* At least one active Nomad server
* **Recommended:** A VictoriaMetrics server containing allocation level resource statistics
* Docker Version: 19.03.8+

## Setup
The user can configure NURD to connect to a containerized SQL Server instance with `docker-compose.yml` or point to another SQL Server instance with `Dockerfile`. See options below for details. 

### Containerized SQL Server Instance
1. `$ git clone git@github.com:Roblox/nurd.git`
2. **Configuration**<br>
    a. **[docker-compose.yml](https://github.com/Roblox/nurd/blob/master/docker-compose.yml)**<br>
        This file contains the necessary login information to create a SQL Server instance. It is necessary to configure the system administrator password and the connection string.<br>
    b. **[etc/nurd/config.json](https://github.com/Roblox/nurd/blob/master/etc/nurd/config.json)**<br>
        This file contains the configuration information for the Nomad server(s) and the VictoriaMetrics server. The default installation contains server addresses for Alpha. Note, any amount of servers can be added to the `Nomad` array.
3. `$ cd nurd`
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
    h. Upload [grafana.json](https://github.com/Roblox/nurd/blob/master/grafana.json) and select `import`<br>


### Another SQL Server Instance
1. `$ git clone git@github.com:Roblox/nurd.git`
2. **Configuration**<br>
    a. **[Dockerfile](https://github.com/Roblox/nurd/blob/master/Dockerfile)**<br>
        This file contains the necessary login information to connect to a separate SQL Server instance. It is necessary to configure the connection string environment variable.<br>
    b. **[etc/nurd/config.json](https://github.com/Roblox/nurd/blob/master/etc/nurd/config.json)**<br>
        This file contains the configuration information for the Nomad server(s) and the VictoriaMetrics server. The default installation contains server addresses for Alpha. Note, any amount of servers can be added to the `Nomad` array.
3. `$ cd nurd`
4. `$ docker build -t nurd .`
5. `$ docker run -dp 8080:8080 nurd`

## Exit
1. `$ docker-compose down` __or__ `$ docker stop`

## Usage
### Grafana Dashboard
From [localhost:3000](http://localhost:3000), or an alternative NURD host address, the user can access the Grafana dashboard. Note, no time series will display until NURD has inserted data into the database. The following parameters are available to query through the dropdown menu:<br>
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
    `http://localhost:8080/jobs`

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
#### Reload Config File
The user can reload the config file without restarting NURD by sending a SIGHUP signal.<br>

`$ kill -S HUP <PID>`<br>

Once the config file has been reloaded and SIGHUP has been sent to NURD, NURD will complete resource aggregation of the addresses in the previous config file before aggregating on the new addresses. 