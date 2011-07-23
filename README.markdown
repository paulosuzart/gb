Intro
=====

gb is a stress test tool based on [Apache Benchmark](http://httpd.apache.org/docs/2.0/programs/ab.html "ab"). It has zero dependencies, so you should be able to build the project and start using it.

Architecture
============

The figure bellow depicts the distributed architeture behind gb:

![gb](http://github.com/paulosuzart/gb/raw/master/arch.jpg)

Note that distributed gb is optional. You can run it in standalone mode.

Using it
========

**Note**: compatible with r57

Before, clone the project and build it:
    
    git clone git@github.com:paulosuzart/gb.git
    gomake
    
You can also use `goinstall github.com/paulosuzart/gb`. `goinstall` will put a compile gb to you `$GOROOT/bin` after clonning the git repo to `$GOROOT/src/pkg/github.com`.

    
Running gb in Master/Workers mode:

Run, say, two Workers:

    ./gb -M worker -H localhost:1978 
    ./gb -M worker -H localhost:1977

They should print something like:

    2011/05/15 13:23:22 Starting in worker mode
    2011/05/15 13:23:22 Setting up a Localworker...
    2011/05/15 13:23:22 Waiting for tasks...

Now you are able to run the Master:
    
    ./gb -M master -W localhost:1978,localhost:1979 -c 2 -n 20 -T 70

Note: Every Worker should be up and running before starting the master

The Master should print something like:

    2011/05/18 00:38:19 Starting in master mode
    2011/05/18 00:38:19 Starting Master...
	2011/05/18 00:38:19 TEST SESSION {1311125723245295000 700000000}
	2011/05/18 00:38:19 2 ProxyWorker(s) may be used by gb
    2011/05/18 00:38:19 Setting up a ProxyWorker for localhost:1978
    2011/05/18 00:38:19 Setting up a ProxyWorker for localhost:1979
    2011/05/18 00:38:19 2 ProxyWorker(s) may be used by gb
    2011/05/18 00:38:19 Tasks distributed. Waiting for summaries...
    2011/05/18 00:38:19 
    =========================================================================
            Test Summary (gb. Version: 0.0.2 beta)
    -------------------------------------------------------------------------                
    Total Go Benchmark Time         | 60.768 milisecs
    Requests Performed              | 20
    Requests Lost                   | 0
    Target supports (reqs/s)        | ~272
    Average Response Time           | 4.7511 milisecs 
    Max Response Time               | 18.089 milisecs
    Min Response Time               | 1.528 milisecs


At the same time Workers should print:

    2011/05/18 00:38:14 Starting in worker mode
    2011/05/18 00:38:14 Waiting for tasks...
    2011/05/18 00:38:19 Task Received from localhost:1970
    2011/05/18 00:38:19 Summary sent to localhost:1970

`netchan import: header:EOF` is a Go log.   

Parameters
==========

Available parameters by now are:

 *   `-c concurrent`. Number of clients to perform the requests.
 *   `-C Content-Type`. Http header to be sent.
 *   `-n requests`. Number of request each client should perform.
 *   `-A username:password`. For Http Basic Authentication.
 *   `-M mode`: standalone, master, worker.
 *   `-H host`: Used for identify the host running gb. No effect in standalone mode. Default is ($hostname):1970.
 *   `-W workers addresses`: Used for distributed gb. Separated by comma.
 *   `-T max time`: Max time in milisecs for gb execution. 
 *   `-t target`. Target http server. The protocol is mandatory.
 *   `-O cookie`. cookie-name=value. A Cookie Header to be added to request.

Licensing?
==========
Eclipse Public License 1.0 (http://opensource.org/licenses/eclipse-1.0.php)


TODO
====
 *   Websocket to report in real time the status of request. A browser will be
welcome.
 *   Test coverage using a go test server to ensure all the options provided by GB.
 *   Timeout for workers and Master. **DONE!** 
 *   Distribute workers using a worker mode for gb. **DONE!**
 *   Cover HTTP POST. **DONE!** Now the usage of `-C` sets the method to POST.
 *   Support for cookies. **DONE!** Need some improvements.
 *   File upload.
 *   POST data file.
 *   Request parameters by csv file
 *   Enable standalone mode again. **DONE!**
 *   Improve netchan.Importer usage in worker mode. **DONE! Now workers keeps the channel open for masters no more than -M. after -M the worker closes the Test session (imported channel) by its own, avoiding holding the dead channel forever.**
