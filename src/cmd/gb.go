package main

import (
	"http"
	"log"
	"time"
	"flag"
	"os"
)

var (
	concurrent = flag.Int("c", 1, "Number of concurrent users emulated. Default 1.")
	requests   = flag.Int("n", 1, "Number of total request to be performed. Default 1.")
	target     = flag.String("t", "http://localhost:8089", "Target to perform the workload.")
)


//Starts concurrent number of workers and waits for everyone terminate. 
//Computes the average time and log it.
func main() {
	flag.Parse()
	log.Print("Starting requests...")

	master := &Master{
		monitor:  make(chan *workSumary),
		ctrlChan: make(chan bool),
	}
	master.BenchMark()

	//wait for the workers to complete after sumarize
	<-master.ctrlChan
	log.Printf("Job done.")

}

type Master struct {
	monitor  chan *workSumary
	ctrlChan chan bool
	workers  map[*Worker]Worker
}

func (m *Master) BenchMark() {
	// starts the sumarize reoutine.
	go m.Sumarize()
	m.workers = map[*Worker]Worker{}
	for c := 0; c < *concurrent; c++ {

		//create a new Worker	
		w := &Worker{
			httpClient: new(http.Client),
			resultChan: m.monitor,
			work:       perform,
			requests:   *requests,
		}
		m.workers[w] = *w
		// a go for the Worker
		go w.Execute()
		// #TODO if a worker get stuck it will never send back the result
		// we need a timout for every worker.
	}

}

func (m *Master) Sumarize() {
	var start, end int64
	start = time.Nanoseconds()
	var avg float64 = 0
	totalSuc := 0
	totalErr := 0
	//totalReqs := *concurrent * *requests

	for result := range m.monitor {
		//remove the worker from master
		m.workers[result.Worker] = m.workers[result.Worker], false

		avg = (result.avg + avg) / 2
		totalSuc += result.sucCount

		//if workers
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
	errCount int
	sucCount int
	avg      float64
	Worker   *Worker
}

//A worker
type Worker struct {
	work       func(*http.Client) (float64, os.Error)
	resultChan chan *workSumary
	httpClient *http.Client
	requests   int
}

// put the avg response time for the executor.
func (w *Worker) Execute() {
	var totalElapsed float64
	totalErr := 0
	totalSuc := 0

	for i := 0; i < w.requests; i++ {
		elapsed, err := w.work(w.httpClient)
		if err == nil {
			totalSuc += 1
			totalElapsed += elapsed
		} else {
			totalErr += 1
		}
	}
	w.resultChan <- &workSumary{
		errCount: totalErr,
		sucCount: totalSuc,
		avg:      totalElapsed / float64(totalSuc),
		Worker:   w,
	}
}

func perform(client *http.Client) (float64, os.Error) {
	start := time.Nanoseconds()

	resp, _, err := client.Get(*target)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Print(err.String())
		return 0, err
	}
	end := time.Nanoseconds()
	total := float64((end - start) / 1000000)

	return total, nil
}
