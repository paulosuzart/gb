// Copyright (c) Paulo Suzart. All rights reserved.
// The use and distribution terms for this software are covered by the
// Eclipse Public License 1.0 (http://opensource.org/licenses/eclipse-1.0.php)
// which can be found in the file epl-v10.html at the root of this distribution.
// By using this software in any fashion, you are agreeing to be bound by
// the terms of this license.
// You must not remove this notice, or any other, from this software.

package main

import (
	"log"
	"strings"
	"netchan"
	"flag"
	"time"
	"template"
)

var (
	concurrent   = flag.Int("c", 1, "Number of concurrent users emulated. Default 1.")
	requests     = flag.Int("n", 1, "Number of total request to be performed. Default 1.")
	target       = flag.String("t", "http://localhost:8089", "Target to perform the workload.")
	unamePass    = flag.String("A", "", "auth-name:password")
	workersAddrs = flag.String("W", "localhost:1977", "The worker Addr")
)

//Creates a serie of workers regarding the gb mode
//for the given master
func produceWorkers(master *Master) (workers []Worker) {
	var wtype string
	createLocalWorkers := func() {
		wtype = "Local"
		workers = make([]Worker, *concurrent)
		for c := 0; c < *concurrent; c++ {
			wk := NewLocalWorker(master.mode, nil)
			wk.SetMasterChan(master.channel)
			go wk.Serve()
			workers[c] = wk
		}

	}
	createProxyWorkers := func() {
		wtype = "Proxy"
		addrs := strings.Split(*workersAddrs, ",", -1)
		workers = make([]Worker, len(addrs))
		for i, addr := range addrs {
			//Try to connect
			wk, err := NewProxyWorker(addr)
			if err != nil {
				log.Panicf("Unable to connect %v Worker\n make sure it is running", addr)
			}
			wk.Serve()
			workers[i] = wk
		}
	}

	switch *master.mode {
	case "standalone":
		createLocalWorkers()
	case "master":
		createProxyWorkers()
	}
	log.Printf("%v %vWorker(s) may be used by gb", len(workers), wtype)
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
	channel      chan WorkSummary //workers reports by WorkSummary
	ctrlChan     chan bool
	runningTasks int
	mode         *string
	exptr        *netchan.Exporter
	summary      *Summary //Master summary 
	done         bool
	session      Session
}

//Every master has its own session.
//A sessions has an Id, that is simply the current nanoseconds.
//It helps workers kill (for worker mode) any dead channel
//imported from finished masters.
type Session struct {
	Id, Timeout int64
}

//The resunting summary of a master
type Summary struct {
	Start, End, Max    int64
	TotalSuc, TotalErr int
	Min                int64
	Avg                float64
	Elapsed            int64
}

func (self *Summary) String() string {
	s := `
=========================================================================
        Test Summary (gb. Version: 0.0.1 alpha)
-------------------------------------------------------------------------                
Total Go Benchmark time         | {Elapsed} miliseconds
Requests performed              | {TotalSuc}
Average response time           | {Avg} miliseconds 
Max Response Time               | {Max} milisecs
Min Response Time               | {Min} milisecs
Requests losts                  | {TotalErr}
`
	t := template.MustParse(s, nil)
	sw := new(StringWritter)
	t.Execute(sw, self)
	return sw.s
}

//In case of timeout, this func is called by gb.go
func (self *Master) Shutdown() {
	if self.done {
		return
	}
	self.done = true
	if *self.mode == "master" {
		self.exptr.Hangup("masterChannel")
	}
	if self.summary.End == 0 {
		self.summary.End = time.Nanoseconds()
		self.summary.Elapsed = (self.summary.End - self.summary.Start) / 1000000
	}
	//log.Print(self.summary)
}

func newSession(timeout int64) Session {
	s := &Session{Id: time.Nanoseconds(), Timeout: timeout}
	return *s
}

//New Master returned. If mode is master, attempts to export the
//master channel for workers.
//Timout is also considere.
func NewMaster(mode, hostAddr *string, timeout int64) *Master {
	log.Print("Starting Master...")
	masterChan := make(chan WorkSummary, 10)
	m := new(Master)
	m.session = newSession(timeout)

	log.Printf("TEST SESSION %v", m.session)
	if *mode == "master" {
		m.exptr = netchan.NewExporter()
		m.exptr.Export("masterChannel", masterChan, netchan.Recv)
		m.exptr.ListenAndServe("tcp", *hostAddr)
	}

	m.channel = masterChan
	//m.ctrlChan = make(chan bool)
	m.mode = mode
	m.summary = &Summary{Min: -1}

	return m

}
//For each client passed by arg, a new worker is created.
//Workers pointers are stored in m.workers to check the end of
//work for each one.
func (m *Master) BenchMark(ctrlChan chan bool) {
	// starts the sumarize reoutine.
	m.ctrlChan = ctrlChan

	u, p := getCredentials()
	newTask := func() (t *Task) {
		t = new(Task)
		t.Host = *target
		t.Requests = *requests
		t.MasterAddr = *hostAddr
		t.User = u
		t.Password = p
		t.Session = m.session
		return
	}

	workers := produceWorkers(m)
	go m.summarize()
	load := *concurrent / len(workers)
	remain := *concurrent % len(workers)
	for _, w := range workers {
		for l := 0; l < load; l++ {
			m.runningTasks += 1
			newTask().Send(w)
		}
	}
	//The remaining work goes for the
	//first worker        
	for r := 0; r < remain; r++ {
		m.runningTasks += 1
		newTask().Send(workers[0])
	}

}

//Read back the workSumary of each worker.
//Calculates the average response time and total time for the
//whole request.
func (self *Master) summarize() {
	log.Print("Tasks distributed. Waiting for summaries...")
	self.summary.Start = time.Nanoseconds()
	for tSummary := range self.channel {
		//remove the worker from master
		self.runningTasks -= 1

		self.summary.Avg = (tSummary.Avg + self.summary.Avg) / 2
		self.summary.TotalSuc += tSummary.SucCount
		self.summary.TotalErr += tSummary.ErrCount

		self.summary.Max = Max(self.summary.Max, tSummary.Max)

		self.summary.Min = Min(self.summary.Min, tSummary.Min)
		//if no workers left 
		if self.runningTasks == 0 {
			self.summary.End = time.Nanoseconds()
			self.summary.Elapsed = (self.summary.End - self.summary.Start) / 1000000
			break
		}

	}

	self.ctrlChan <- true
}
