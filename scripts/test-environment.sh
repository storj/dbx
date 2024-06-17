#!/usr/bin/env bash

set -euxo pipefail

function retry() {
  set +e
  counter=0
  until "$@"; do
    sleep 1
    [[ counter -eq 10 ]] && echo "Failed!" && exit 1
    echo "Trying again. Try #$counter"
    ((counter++))
  done
  set -e
}

if [[ ${STORJ_TEST_POSTGRES:=""} != "omit" ]]; then
   echo "Starting Postgresql"
   service postgresql start
   retry psql -U postgres -c 'create database testdb';
   export STORJ_TEST_POSTGRES="postgres://postgres@localhost/testdb?sslmode=disable"
fi

if [[ ${STORJ_TEST_COCKROACH:=""} != "omit" ]]; then
  echo "Starting Cockroach"
   cockroach start-single-node \
     --insecure \
     --store=type=mem,size=4GiB \
     --listen-addr=localhost:26257 \
     --http-addr=localhost:8086 \
     --cache 1024MiB \
     --max-sql-memory 1024MiB >/dev/null 2>&1 &
   COCKROACH_PID=$!
   echo "Cockroach db is started $COCKROACH_PID"
   retry cockroach sql --insecure --host=localhost:26257 -e 'create database testcockroach;'
   #TODO: tests are not ready yet, as generated PKs are not handled very well...
   #export STORJ_TEST_COCKROACH='cockroach://root@localhost:26257/testcockroach?sslmode=disable'
fi

if [[ ${STORJ_TEST_SPANNER:=""} != "omit" ]]; then
   spanner_emulator --host_port 127.0.0.1:9010 2>&1 &
   SPANNER_PID=$!
   export STORJ_TEST_SPANNER=spanner://127.0.0.1:9010?emulator=true
fi

set +e
"$@"
RESULT=$?
echo "Killing Cockroach"
kill $COCKROACH_PID
echo "Stopping Postgres"
service postgresql stop
echo "Killing Spanner emulator"
kill $SPANNER_PID
exit $RESULT