Intro
=====

gb is a stress test tool based on Apache Benchmark. It has zero dependencies, so you should be able to build the projet and start using it.


Running gb in Master/Workers mode:

Run, say, two Workers
    ./gb -M worker -H localhost:1978 
    ./gb -M worker -H localhost:1977

They should print something like:

    2011/05/15 13:23:22 Starting in worker mode
    2011/05/15 13:23:22 Setting up a Localworker...
    2011/05/15 13:23:22 Waiting for tasks...

Now you are able to run the Master:

    ./gb -M master -c 5 -n 20 -W localhost:1977,localhost:1978 -t http://localhost:8089

Note: Every Worker should be up and running before starting the master

The Master should print something like:

    2011/05/15 13:26:50 Starting in master mode
    2011/05/15 13:26:50 Starting Master...
    2011/05/15 13:26:50 Setting up a ProxyWorker
    2011/05/15 13:26:50 Setting up a ProxyWorker
    2011/05/15 13:26:50 2 ProxyWorkers will be used by gb
    2011/05/15 13:26:50 Tasks distributed. Waiting for summaries...
    2011/05/15 13:26:50 Total Go Benchmark time 115 miliseconds.
    2011/05/15 13:26:50 100 requests performed. Average response time 5.504540625e+06 miliseconds.
    2011/05/15 13:26:50 0 requests lost.

At the same time Workers should print:

    2011/05/15 13:26:50 Task Received from localhost:9393
    2011/05/15 13:26:50 Task Received from localhost:9393
    2011/05/15 13:26:50 Summary sent to localhost:9393
    2011/05/15 13:26:50 Summary sent to localhost:9393
    2011/05/15 13:26:50 netchan import: header:EOF
    2011/05/15 13:26:50 netchan import: header:EOF

`netchan import: header:EOF` is a Go log.   

Parameters
==========

Available parameters by now are:

 *   `-t` target. It may change, bu represents the target http server.
 *   `-c` concurrent. Number of clients to perform the requests.
 *   `-n` requests. Number of request each client should perform.
 *   `-A` username:password. For Http Basic Authentication.
 *   `-M` mode: standalone, master, worker.
 *   `-H` host: Used for identuify the host running gb. No effect in standalone mode.
 *   `-W` workers addresses: Used for distributed gb. 

Licencing?
==========
None yet.

paulosuzart@gmail.com

TODO
====
.Websocket to report in real time the status of request. A browser will be
welcome.
.Timeout for workers and Master
.Distribute workers using a worker mode for gb. DONE!
