#!/bin/bash

/usr/bin/rtlamr -unique=true -msgtype=scm -format=json |\
    tee -a /opt/powermeter/backup.json |\
    tee >(/opt/powermeter/powermeter-sqlite serve -cache /opt/powermeter/power.gob -db /opt/powermeter/power.db -meter 18011759 -http 0.0.0.0:5000 ) |\
    /opt/powermeter/powermeter-influxdb -host http://localhost:8086 -cache /opt/powermeter/cache.gob
