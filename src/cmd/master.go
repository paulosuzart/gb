package main

import (
	"log"
	"strings"
	"netchan"
	"flag"
	"time"
)

var (
	concurrent   = flag.Int("c", 1, "Number of concurrent users emulated. Default 1.")
	requests     = flag.Int("n", 1, "Number of total request to be performed. Default 1.")
	target       = flag.String("t", "http://localhost:8089", "Target to perform the workload.")
	unamePass    = flag.String("A", "", "auth-name:password")
	workersAddrs = flag.String("W", "localhost:1977", "The worker Addr")
)

//Creates a serie of workers regarding the gb mode
func getWorkers(mode string) (workers []Worker) {
	var wtype string
	if mode == "standalone" {
                wtype = "Local"
		workers = make([]Worker, *concurrent)
		for c := 0; c < *concurrent; c++ {
			workers[c] = NewLocalWorker(nil, nil)
		}
	} else if mode == "master" {
                wtype = "Proxy"
		addrs := strings.Split(*workersAddrs, ",", -1)
		workers = make([]Worker, len(addrs))
		for i, addr := range addrs {
			//Try to connect
			wk, err := NewProxyWorker(addr)
			if err != nil {
				log.Panicf("Unable to connect %v Worker", addr)
			}

			workers[i] = wk
		}
	}
        log.Printf("%v %vWorkers will be used by gb", len(workers), wtype)        
	return

}
//Extracts credentials from command line arguments
func getCredentials() (u, p string) {
	if *unamePass == "" {
		return
	}
	authData := strings.Split(*unamePass, ":", 2)

	if len(authData) != 2 {
		log.Panic("No valid credentials found in -A argument")
	}
	u = authData[0]
	p = authData[1]
	return
}


//Represents this master.
type Master struct {
	channel         chan WorkSummary
	ctrlChan        chan bool
	runningWorkers  int
	mode            string
}


func NewMaster(mode, hostAddr *string) (m *Master) {
	log.Print("Starting Master...")
	masterChan := make(chan WorkSummary, 10)

	if *mode == "master" {
		e := netchan.NewExporter()
		e.Export("masterChannel", masterChan, netchan.Recv)
		e.ListenAndServe("tcp", *hostAddr)
	}

	m = &Master{
		channel:  masterChan,
		ctrlChan: make(chan bool),
		mode:     *mode,
	}
	return

}
//For each client passed by arg, a new worker is created.
//Workers pointers are stored in m.workers to check the end of
//work for each one.
func (m *Master) BenchMark() {
	// starts the sumarize reoutine.

	// #TODO if a worker get stuck it will never send back the result
	// we need a timout for every worker.
	u, p := getCredentials()
	newTask := func() (t *Task) {
		t = new(Task)
		t.Host = *target
		t.Requests = *requests
		t.MasterAddr = *hostAddr
		t.User = u
		t.Password = p
		return
	}

	workers := getWorkers(m.mode)
	load := *concurrent / len(workers)
	remain := *concurrent % len(workers)
	for _, w := range workers {
		for l := 0; l < load; l++ {
			newTask().Send(w)
			m.runningWorkers += 1
		}
	}
	//The remaining work goes for the
	//first worker         
	for r := 0; r < remain; r++ {
		newTask().Send(workers[0])
		m.runningWorkers += 1
	}
	go m.Summarize()
        
}

//Read back the workSumary of each worker.
//Calculates the average response time and total time for the
//whole request.
func (m *Master) Summarize() {
	log.Print("Tasks distributed. Waiting for summaries...")
	var start, end int64
	var avg float64 = 0
	totalSuc := 0
	totalErr := 0

	start = time.Nanoseconds()
	for result := range m.channel {
		//remove the worker from master
		m.runningWorkers -= 1
		avg = (result.Avg + avg) / 2
		totalSuc += result.SucCount
		totalErr += result.ErrCount

		//if no workers left 
		if m.runningWorkers == 0 {
			end = time.Nanoseconds()
			break
		}

	}

	log.Printf("Total Go Benchmark time %v miliseconds.", (end-start)/1000000)
	log.Printf("%v requests performed. Average response time %v miliseconds.", totalSuc, avg)
	log.Printf("%v requests lost.", totalErr)
	m.ctrlChan <- true

}
