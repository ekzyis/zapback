#!/usr/bin/env bash

PID=$(pidof zapback)
DIRS="db/ env/ lib/ lightning/ pages/ public/ server/ main.go"

set -e

echo ":: remote port forwarding for zapback-dev.ekzy.is ::"
ssh -fnNR 8888:localhost:4444 zapback-dev.ekzy.is
echo

function restart_server() {
  set +e
  [[ -z "$PID" ]] || kill -15 $PID
  ENV=development make build -B
  set -e
  ./zapback 2>&1 &
  PID=$(pidof zapback)
}

function restart() {
  restart_server
  # give server time start listening for connections
  sleep 1
  date +%s.%N > public/__livereload
}

function cleanup() {
    rm -f public/__livereload
    [[ -z "$PID" ]] || kill -15 $PID
}
trap cleanup EXIT

restart

while inotifywait -r -e modify $DIRS; do
  restart
done
