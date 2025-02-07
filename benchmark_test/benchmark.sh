#!/bin/bash
# usage: ./benchmark.sh <no_of_servers> <no_of_threads>

# Default number of servers is 3
BASE_PORT=8100
NUM_SERVERS=${1:-3}  # If the first argument is provided, use it; otherwise, default to 3
THREADS=${2:-1}  # If the first argument is provided, use it; otherwise, default to 3
WAIT_TIME=5  # Time to wait in seconds before running the test
PIDS=()      # Array to store server process IDs

# Function to terminate all background servers on script exit
cleanup() {
  # Pattern to match processes (e.g., processes related to go build)
  MATCH_PATTERN="/tmp/go-build"

  echo "Killing processes with the pattern: $MATCH_PATTERN"

  # Find and kill processes matching the pattern
  ps aux | grep "$MATCH_PATTERN" | grep -v grep | awk '{print $2}' | while read pid; do
    echo "Killing process PID: $pid"
    kill -9 $pid
  done

  # Additionally, stop all the servers that we have stored PIDs for
  echo "Stopping all servers..."
  for pid in "${PIDS[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "Killing server process $pid..."
      kill -9 "$pid" 2>/dev/null   # Force kill the server processes
    else
      echo "Server process $pid already stopped."
    fi
  done
  echo "All matching processes and servers have been killed."
}
trap cleanup EXIT   # Calls cleanup function on script exit (including Ctrl+C)

# Start the servers
for ((i=0; i<NUM_SERVERS; i++)); do
  PORT=$((BASE_PORT + i))
  echo "Starting server on port $PORT..."
#  ulimit -n 65536                          # Increase the open file limit
  go run server/server.go -port="127.0.0.1:$PORT" &   # Start server in background
  PIDS+=($!)                                # Store the server's PID

  sleep 0.5
done

echo "All servers started. Waiting for $WAIT_TIME seconds before running benchmark..."
sleep $WAIT_TIME

# Run the test
echo "Running benchmark..."
go test -run a$ -bench "Gol/512x512x1000-$THREADS" -timeout 100s -threads $THREADS
# go test -v -run TestGol/-1$

# Cleanup will be triggered here by the `trap` when the script ends
