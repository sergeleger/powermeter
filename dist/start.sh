#!/bin/bash

/usr/bin/rtlamr -unique=true -msgtype=scm -format=json |\
    tee -a /opt/powermeter/backup.json |\
    /opt/powermeter/powermeter -host http://localhost:8086 -cache /opt/powermeter/cache.gob
