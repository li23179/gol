# Conway's Game of Life

## Pre-requisite 

Please install these libraries / environments, AWS node is not necessary.
- [anaconda](https://www.anaconda.com/download)
- [Go](https://go.dev/doc/install)
- [AWS](https://aws.amazon.com/free/?trk=d5254134-67ca-4a35-91cc-77868c97eedd&sc_channel=ps&ef_id=Cj0KCQiAuou6BhDhARIsAIfgrn76pNL3xYzPc5pSQmvSDhgOGS4xrOOh7ogFVgJ9IVFIs6OMwDxKmqUaAq4eEALw_wcB:G:s&s_kwcid=AL!4422!3!433803620858!e!!g!!aws!1680401428!67152600164&gclid=Cj0KCQiAuou6BhDhARIsAIfgrn76pNL3xYzPc5pSQmvSDhgOGS4xrOOh7ogFVgJ9IVFIs6OMwDxKmqUaAq4eEALw_wcB&all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc&awsf.Free%20Tier%20Types=*all&awsf.Free%20Tier%20Categories=*all)

## Introduction

This repository features two versions of **Conway's Game of Life**, collaboratively implemented by [**Gordon Wai Hin Kam**](https://github.com/li23179) and [**Ivan (Ka Ho Leung)**](https://github.com/nm22031) in [Go](https://go.dev/).

**Conway's Game of Life** is a cellular automaton devised by the British mathematician John Horton Conway in 1970. 

A user can only interact with **Game of Life** by creating an initial configuration and observing its evolution.

## Rules
**Conway's Game of Life** is implemented in a closed domain by reading an `nxn.pgm` file, which sets up the initial state based on the provided image grid.

At each cell in matrix update in time the following transitions may occur to create the next evolution of the domain:

- any **live** cell with fewer than two live neighbours dies
- any **live** cell with two or three live neighbours is unaffected
- any **live** cell with more than three live neighbours dies
- any **dead** cell with exactly three live neighbours becomes alive

## Implementation

**Parallel Implementation**: This version utilizes multithreading to enhance performance by distributing the workload across multiple CPU cores via [Channels](https://gobyexample.com/channels) and [Mutex Lock](https://gobyexample.com/mutexes) to achieve **free** [Race Condition](https://en.wikipedia.org/wiki/Race_condition).

**Parallel-Distributed Implementation**: This version extends the scalability by distributing the processing across multiple nodes via [RPC (Remote Procedure Call)](https://en.wikipedia.org/wiki/Remote_procedure_call) in a network and as well as each node distributing the workload across multiple threads.

## Features of Game of Life

### Key Press Event

When you run the game, you can press the following keys during the execution of Game of life:

- Press `s` will save the state of world and output the `nxnxt.pgm` file in the [image](/parallel/) (depends on which version you are running) directory, where `t` is the number of turns when you press `s`.

- Press `q` will terminate the game and output the `nxnxt.pgm` file, similar as above.

- Press `p` will pause the state of world, and press `p` again will resume the game. Note, you can press `s` and `q` when the game state is paused.


- **(Only works for Parallel-Distributed System)** Press `k` will gracefully shutdown all the components in distributed system.

### Alive Cells Ticker Event
When you run the game, the CLI will output the current number of alive cells and turns for **every 2 seconds**.

## Running Game of Life

### Parallel Version

Go to [parallel](/parallel/) directory.

If you want to run the Parallel version of game of life, run the following command on terminal: 
```
go run .
```

If you want to run the test of game of life, run the following command on terminal:
```
go test -v -race
```

`-v` is a flag which tells the Go testing framework to print more detailed output about the tests that are being run.

`-race` is a flag which detects any race condition.

### Parallel-Distrubuted Version

Go to [parallel_distributed](/parallel_distributed/) directory.

This is getting tricky compared to parallel version.

You can run it either on [AWS](https://aws.amazon.com/free/?trk=d5254134-67ca-4a35-91cc-77868c97eedd&sc_channel=ps&ef_id=Cj0KCQiAuou6BhDhARIsAIfgrn76pNL3xYzPc5pSQmvSDhgOGS4xrOOh7ogFVgJ9IVFIs6OMwDxKmqUaAq4eEALw_wcB:G:s&s_kwcid=AL!4422!3!433803620858!e!!g!!aws!1680401428!67152600164&gclid=Cj0KCQiAuou6BhDhARIsAIfgrn76pNL3xYzPc5pSQmvSDhgOGS4xrOOh7ogFVgJ9IVFIs6OMwDxKmqUaAq4eEALw_wcB&all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc&awsf.Free%20Tier%20Types=*all&awsf.Free%20Tier%20Categories=*all) nodes or locally.

### Broker

First, we need to run the [broker](/parallel_distributed/broker/).

You can add `-port` flag or not. If you add port flag then replace `<port>` with the port that broker want to listen on

Run the following command in terminal:
```
go run broker/broker.go -port="<port>"
```

`port` flag is the port that **broker** listen on. If not specified, the default port number is `8030`.

### Server

Second, we need to run the [server](/parallel_distributed/server/).

You can add `-broker` flag or not. If you add port flag then replace `<broker-port>` with the port that broker listened on. Similar to `port` flag.

Run the following command in terminal:
```
go run server/server.go -broker="<broker-port>" -port="<port>"
```

`broker` flag is the port address that **broker** listen on, please enter the correct broker address as you enter above. If not specified, the default port number is `8030`.

`port` flag is the port that **server** listen on. If not specified, the default port number is `8050`.

Note: you can run more than one server by changing the `port` flag number.

### Local Controller

Lastly, if you want to run the game locally then run the following command on terminal:
```
go run .
```
if you want to run the test then run the following command on terminal:
```
go test -v -race
```

Note: It is normal that the SDL window shows a black screen for Parallel Distributed System since we have not implement the events to sdl window.

## Benchmark Test

This section provides details on the benchmark tests conducted to evaluate the performance and efficiency of implementations.

### Objectives
The primary objectives of our benchmark tests include:

- **Performance Evaluation**: Measure the execution time and processing speed under various computational loads.
- **Scalability Assessment**: Determine how effectively the software can scale up with increasing input size.
- **Resource Utilization**: Analyze CPU and memory usage during peak operation conditions.

In the [benchmark_test](/benchmark_test/) directory, [benchmark_test.go](/benchmark_test/benchmark_test.go) and [halo.go](/benchmark_test/halo.go) are benchmark test. You can simply copy these files to the version of Game of Life (gol).

[benchmark_test.go](/benchmark_test/benchmark_test.go) is used to test the performance of parallel component.

[halo.go](/benchmark_test/halo.go) is used to test the performance of direct halo exchange between servers.

Copy the Benchmark function you want to test in the version of Game of Life.
```
cp benchmark_test/<Benchmark.go> <version_of_gol_directory>
```

Then, run the following command:

This will repeat each sub-benchmark 5 times, but each result will be reported individually.

```
go test -run ^$ -bench . -benchtime 1x -count 5 | tee results.out
```

We can now use our benchstat library to convert raw benchmark output to a 'Comma Separated Values' (CSV) file.

```
go run golang.org/x/perf/cmd/benchstat -format csv results.out | tee results.csv
```

There are some python script to plot the graph of performance.

Similarly, copy the python plot program you want to test in the version of Game of Life.
```
cp benchmark_test/<Benchmark.py> <version_of_gol_directory>
```

Run the following command:
```
python <Benchmark.py>
```

### Report

We have written a report to analyse the performance in each component and also the functionality of Game of Life we have implemented in [**Report**](/report.pdf).
___
&copy; 2024 Gordon Wai Hin Kam &amp; Ivan Leung Ka Ho. All rights reserved. 