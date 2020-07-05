# Nomad Usage Resource Dashboard (NURD)
NURD is a dashboard which aggregates and displays CPU and memory resource usage for each job running through specified Hashicorp Nomad servers. The dashboard also displays resources requested by each job, which can be used with resource usage to calculate waste and aid capacity planning. 

## Prerequisites
* At least one active Nomad server
* **Recommended:** A VictoriaMetrics server containing allocation level resource statistics

## Setup
1. **Configuration File**<br>
    a. **nurd/config.json**<br>
        This file contains the configuration information for the Nomad server(s) and the VictoriaMetrics server. Note, any amount of servers can be added to the `Nomad` array.
        ```
        {
            "VictoriaMetrics": {
                "URL":      URL for VictoriaMetrics server 
                "Port":     Port for VictoriaMetrics server
            },
            "Nomad": [
                {
                    "URL":      URL for Nomad server
                    "Port":     Port for Nomad server
                }
            ]
        }
        ```
