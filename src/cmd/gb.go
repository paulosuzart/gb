package main

import (
	"log"
	"time"
	"flag"
	"github.com/paulosuzart/gb/gbclient"
	"strings"
)

var (
	concurrent = flag.Int("c", 1, "Number of concurrent users emulated. Default 1.")
	requests   = flag.Int("n", 1, "Number of total request to be performed. Default 1.")
	target     = flag.String("t", "http://localhost:8089", "Target to perform the workload.")
	unamePass  = flag.String("A", "", "auth-name:password")
	uname      = ""
	passwd     = ""
	basicAuth  = false
)


//Starts concurrent number of workers and waits for everyone terminate. 
//Computes the average time and log it.
func main() {
	flag.Parse()
	log.Print("Starting requests...")

	authData := strings.Split(*unamePass, ":", 2)
	if len(authData) == 2 {
		uname = authData[0]
		passwd = authData[1]
		basicAuth = true
	}

	master := &Master{
		monitor:  make(chan *workSumary, 10),
		ctrlChan: make(chan bool),
	}
	master.BenchMark()

	//wait for the workers to complete after sumarize
	<-master.ctrlChan
	log.Printf("Job done.")

}

//Represents this master.
type Master struct {
	monitor  chan *workSumary
	ctrlChan chan bool
	workers  map[*Worker]Worker
}

//For each client passed by arg, a new worker is created.
//Workers pointers are stored in m.workers to check the end of
//work for each one.
func (m *Master) BenchMark() {
	// starts the sumarize reoutine.
	go m.Sumarize()
	m.workers = map[*Worker]Worker{}

	for c := 0; c < *concurrent; c++ {

		//create a new Worker	
		var w Worker
		w.httpClient = gbclient.NewHTTPClient(*target, "GET")

		if basicAuth {
			w.httpClient.Auth(uname, passwd)
		}
		w.resultChan = m.monitor
		w.requests = *requests

		m.workers[&w] = w

		// a go for the Worker
		go w.Execute()
		// #TODO if a worker get stuck it will never send back the result
		// we need a timout for every worker.
	}

}

//Read back the workSumary of each worker.
//Calculates the average response time and total time for the
//whole request.
func (m *Master) Sumarize() {
	var start, end int64
	var avg float64 = 0
	totalSuc := 0
	totalErr := 0

	start = time.Nanoseconds()
	for result := range m.monitor {
		//remove the worker from master
		m.workers[result.Worker] = m.workers[result.Worker], false

		avg = (result.avg + avg) / 2
		totalSuc += result.sucCount
		totalErr += result.errCount

		//if no workers left 
		if len(m.workers) == 0 {
			end = time.Nanoseconds()
			break
		}

	}

	log.Printf("Total Go Benchmark time %v miliseconds.", (end-start)/1000000)
	log.Printf("%v requests performed. Average response time %v miliseconds.", totalSuc, avg)
	log.Printf("%v requests lost.", totalErr)
	m.ctrlChan <- true

}

//Reported by the worker through resultChan
type workSumary struct {
	errCount int     //total errors
	sucCount int     //total success
	avg      float64 //average response time
	Worker   *Worker //given worker
}

//A worker
type Worker struct {
	//The work summary is sent to the master through
	//this channel
	resultChan chan *workSumary

	//The actual gbclient.HTTPClient to which compute 
	//the request time
	httpClient *gbclient.HTTPClient

	//Number of requests to be performed
	requests int
}

// put the avg response time for the executor.
func (w *Worker) Execute() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			log.Print("Worker died")
			w.resultChan <- &workSumary {Worker : w}
		}
	}()

	var totalElapsed int64
	totalErr := 0
	totalSuc := 0

	//perform w.request times the request
	for i := 0; i < w.requests; i++ {
		start := time.Nanoseconds()
		_, err := w.httpClient.DoRequest()
		elapsed := (time.Nanoseconds() - start)
		if err == nil {
			totalSuc += 1
			totalElapsed += elapsed
		} else {
			totalErr += 1
		}
	}

	var sumary workSumary
	sumary.errCount = totalErr
	sumary.sucCount = totalSuc
	sumary.avg = float64(totalElapsed / int64(totalSuc))
	sumary.Worker = w

	w.resultChan <- &sumary

}
