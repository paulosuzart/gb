package main

import (
	"log"
	"time"
	"flag"
	//	"github.com/paulosuzart/gb/gbclient"
	"strings"
	"netchan"
)

var (
	concurrent = flag.Int("c", 1, "Number of concurrent users emulated. Default 1.")
	requests   = flag.Int("n", 1, "Number of total request to be performed. Default 1.")
	target     = flag.String("t", "http://localhost:8089", "Target to perform the workload.")
	unamePass  = flag.String("A", "", "auth-name:password")
	basicAuth  = false
	mode       = flag.String("M", "standalone", "standalone, master, worker")
	workerAddr = flag.String("W", "localhost:9397", "The worker Addr")
	masterAddr = flag.String("H", "localhost:9393", "The master Addr")
)

func parseCredentials() (u, p string) {
	if *unamePass == "" {
		return
	}
	authData := strings.Split(*unamePass, ":", 2)

	if len(authData) != 2 {
		log.Fatal("No valid credentials found in -A argument")
	}
	u = authData[0]
	p = authData[1]
	return
}

func main() {

	flag.Parse()

	if *mode == "master" || *mode == "standalone" {
		initMaster()
	} else if *mode == "worker" {
		initWorker()
	}
}


func initWorker() {
	log.Print("Starting worker...")
	NewLocalWorker(true, *workerAddr)
}

//Starts concurrent number of workers and waits for everyone terminate. 
//Computes the average time and log it.
func initMaster() {

	log.Print("Starting Master...")
	masterChan := make(chan WorkSummary, 10)
	if *mode == "master" {
		e := netchan.NewExporter()
		e.Export("masterChannel", masterChan, netchan.Recv)
		e.ListenAndServe("tcp", *masterAddr)
	}

	master := &Master{
		channel:  masterChan,
		ctrlChan: make(chan bool),
	}
	master.BenchMark()

	//wait for the workers to complete after summarize
	<-master.ctrlChan
	log.Printf("Job done.")

}

//Represents this master.
type Master struct {
	channel  chan WorkSummary
	ctrlChan chan bool
	workers  int
}

//For each client passed by arg, a new worker is created.
//Workers pointers are stored in m.workers to check the end of
//work for each one.
func (m *Master) BenchMark() {
	// starts the sumarize reoutine.
	go m.Sumarize()

	for c := 0; c < *concurrent; c++ {

		//create a new Worker	
		var w Worker

		if *mode == "master" {
			w = NewProxyWorker(*workerAddr)
		} else if *mode == "standalone" {
			w = NewLocalWorker(false, "")
		}

		var t *Task = new(Task)
		t.Host = *target
		t.Requests = *requests
		t.MasterAddr = *masterAddr
		t.User, t.Password = parseCredentials()

		m.workers += 1

		// #TODO if a worker get stuck it will never send back the result
		// we need a timout for every worker.
		t.Send(w)
	}

}

//Read back the workSumary of each worker.
//Calculates the average response time and total time for the
//whole request.
func (m *Master) Sumarize() {
	log.Print("Tasks distributed. Waiting for summaries...")
	var start, end int64
	var avg float64 = 0
	totalSuc := 0
	totalErr := 0

	start = time.Nanoseconds()
	for result := range m.channel {
		//remove the worker from master
		m.workers -= 1
		avg = (result.Avg + avg) / 2
		totalSuc += result.SucCount
		totalErr += result.ErrCount

		//if no workers left 
		if m.workers == 0 {
			end = time.Nanoseconds()
			break
		}

	}

	log.Printf("Total Go Benchmark time %v miliseconds.", (end-start)/1000000)
	log.Printf("%v requests performed. Average response time %v miliseconds.", totalSuc, avg)
	log.Printf("%v requests lost.", totalErr)
	m.ctrlChan <- true

}
