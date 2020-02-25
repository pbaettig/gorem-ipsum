#!/bin/bash
set -e

while true; do
    curl -Ssf -o /dev/null http://localhost:8080/health &
    curl -Ssf -o /dev/null http://localhost:8080/health/history &
    curl -Ssf -o /dev/null http://localhost:8080/config/health &
    curl -Ssf -o /dev/null http://localhost:8080/ &
    curl -Ssf -o /dev/null http://localhost:8080/count &
    curl -Ssf -o /dev/null http://localhost:8080/info &
    curl -Ssf -o /dev/null http://localhost:8080/http/get?url=http://httpbin.org/get &

    sleep 2
done