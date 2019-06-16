#!/usr/bin/env bash

docker run -d --name influxdb -e INFLUXDB_REPORTING_DISABLED=true -p 8086:8086 -v `pwd`/influxdb:/var/lib/influxdb influxdb:1.7.6-alpine
