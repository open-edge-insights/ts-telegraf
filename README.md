# ts-telegraf

This repo hosts Telegraf service of Edge Insights Software


1. Telegraf will be started by default with the following command based on prod or dev mode as part of ‘telegraf_start.py’:
	```
	$ telegraf -config=/etc/Telegraf/<AppName>/<AppName>.conf
	$ telegraf -config=/etc/Telegraf/<AppName>/<AppName>_devmode.conf
	```
2. In case, user wants to use additional configuration, then user needs to pass the whole command to start the Telegraf in docker-compose.yml file as following:
	```
	command: ["telegraf -config=/etc/Telegraf/<AppName>/<AppName>.conf -config-directory=/etc/Telegraf/<AppName>/telegraf.d"]
	```

**Note**: above two points is applicable for adding multiple telegraf instance.

## Adding multiple telegraf instance:
* User can add multiple instance of telegraf like ia_telegraf1,ia_telegraf2 etc.

* Create a directory inside [config](./config) directory with AppName
  ```
  $ mkdir ./config/<Appname>
  $ cd ./config/<Appname>
  ```
  keep the main configuration files \<AppName>.conf and \<AppName>_devmode.conf in the previously created directory.

All contents inside [config](./config) will be copied inside docker image in /etc/Telegraf in times of docker image building.

**Note**: It's been practice followed by many users, to keep the configuration in a modular way. One of the way to achieve the same could be keeping the additional configuration inside [config](./config)/\<AppName>/telegraf.d. For example:

create a directory ‘telegraf.d’ inside [config](./config)/\<AppName> :
   ```
   $ mkdir config/<AppName>/telegraf.d
   $ cd config/<AppName>/telegraf.d
  ```
   keep additional configuration files inside that directory.


* For adding a new instance user needs to define the telegraf service in [docker-compose.yml](./docker-compose.yml).

	Example:
	```
	  ia_telegraf1:
	    depends_on:
	      - ia_common
	    build:
	      context: $PWD/../Telegraf
	      dockerfile: $PWD/../Telegraf/Dockerfile
	      args:
		EIS_VERSION: ${EIS_VERSION}
		EIS_UID: ${EIS_UID}
		TELEGRAF_VERSION: ${TELEGRAF_VERSION}
		DOCKER_REGISTRY: ${DOCKER_REGISTRY}
	    container_name: ia_telegraf1
	    hostname: ia_telegraf1
	    network_mode: host
	    image: ${DOCKER_REGISTRY}ia_telegraf:${EIS_VERSION}
	    restart: unless-stopped
	    ipc: "none"
	    read_only: true
	    command: ["telegraf -config=/etc/Telegraf/Telegraf1/Telegraf1.conf -config-directory=/etc/Telegraf/Telegraf1/telegraf.d"]
	    environment:
	      AppName: "Telegraf1"
	      InfluxDbAppName: "InfluxDBConnector"
	      CertType: ""
	      DEV_MODE: ${DEV_MODE}
	      no_proxy: ${eis_no_proxy},${ETCD_HOST}
	      NO_PROXY: ${eis_no_proxy},${ETCD_HOST}
	      ETCD_HOST: ${ETCD_HOST}
	      MQTT_BROKER_HOST: '127.0.0.1'
	      INFLUX_SERVER: '127.0.0.1'
	      INFLUXDB_PORT: $INFLUXDB_PORT
	      ETCD_PREFIX: ${ETCD_PREFIX}
	    user: ${EIS_UID}
	    volumes:
	      - "vol_temp_telegraf:/tmp/"
	    secrets:
	      - ca_etcd
	      - etcd_InfluxDBConnector_cert
	      - etcd_InfluxDBConnector_key
	```
* Telegraf Instance can be configured with pressure point data ingestion. In the following example, the MQTT input plugin of Telegraf is configured to read pressure point data and stores into ‘point_pressure_data’ measurement.

	```
	# # Read metrics from MQTT topic(s)
	[[inputs.mqtt_consumer]]
	#   ## MQTT broker URLs to be used. The format should be scheme://host:port,
	#   ## schema can be tcp, ssl, or ws.
		servers = ["tcp://localhost:1883"]
	#
	#   ## MQTT QoS, must be 0, 1, or 2
	#   qos = 0
	#   ## Connection timeout for initial connection in seconds
	#   connection_timeout = "30s"
	#
	#   ## Topics to subscribe to
		topics = [
		"pressure/simulated/0",
		]
		name_override = "point_pressure_data"
		data_format = "json"
	#
	#   # if true, messages that can't be delivered while the subscriber is offline
	#   # will be delivered when it comes back (such as on service restart).
	#   # NOTE: if true, client_id MUST be set
		persistent_session = false
	#   # If empty, a random client ID will be generated.
		client_id = ""
	#
	#   ## username and password to connect MQTT server.
		username = ""
		password = ""

	```


* To start the mqtt-publisher with pressure data,
	```sh
	$ cd ../tools/mqtt-publisher/
	$ ./build.sh && ./publisher.sh --name publisher_pressure --pressure 10:30
	```
please refer [tools/mqtt-publisher/README.md](../tools/mqtt-publisher/README.md)
