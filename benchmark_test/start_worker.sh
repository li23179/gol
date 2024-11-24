#!/bin/bash
# usage: ./start_worker.sh <number_of_workers>

# Check if the number of workers is provided as an argument
if [ -z "$1" ]; then
  echo "Usage: $0 <number_of_workers>"
  exit 1
fi

# Set the base port number
base_port=8090

# Start the servers
for i in $(seq 1 $1)
do
  port=$((base_port + i - 1))
  bash -c "ulimit -n 65536 && go run server/server.go -port=\"127.0.0.1:$port\"" &
done
