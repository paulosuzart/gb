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
	"netchan"
	"time"
	"os"
	"sync"
)

//Represents a set of request to be performed
//against Task.Host        
type Task struct {
	Host, User, Password string
	Requests, Id         int
	MasterAddr           string
	Session              int64
}

//Reported by the worker through resultChan
type WorkSummary struct {
	ErrCount int     //total errors
	SucCount int     //total success
	Avg      float64 //average response time
	Max, Min int64   //the slowest requiest
}


//Put t to w.Channel()        
func (self *Task) Send(w Worker) {
	w.Channel() <- *self
}


//The worker interface
type Worker interface {
	//Should return the input channel to
	//interact with Worker
	Channel() chan Task
	Serve()
}


//A local workers is used in standalone mode
//as well as in worker mode.
type LocalWorker struct {
	//the Worker input channel to
	//receive tasks
	channel       chan Task
	masterChannel chan WorkSummary
	mode          *string
	//ctrlChan      chan bool
}


//Worker interface implemented:w
func (self *LocalWorker) Channel() chan Task {
	return self.channel
}

func (self *LocalWorker) SetMasterChan(c chan WorkSummary) {
	self.masterChannel = c
}


//Creates a new LocalWorker. If export is true, than
//the LocalWorker exports its input channel in the network address
//provided by workerAddr        
func NewLocalWorker(mode, hostAddr *string) (w *LocalWorker) {
	w = new(LocalWorker)
	//w.ctrlChan = make(chan bool)
	w.channel = make(chan Task, 10)
	w.mode = mode
	//exports the channels
	if *mode == "worker" {
		e := netchan.NewExporter()
		e.Export("workerChannel", w.channel, netchan.Recv)
		e.ListenAndServe("tcp", *hostAddr)
	}
	return
}

var _sessions map[int64]chan WorkSummary = make(map[int64]chan WorkSummary)
var mu *sync.RWMutex = new(sync.RWMutex)

//Helper function to import the Master channel from masterAddr
func importMasterChan(masterAddr string, session int64) (c chan WorkSummary) {
	mu.Lock()
	defer mu.Unlock()
	if c, present := _sessions[session]; present {
		log.Printf("cached Session %v", session)
		return c
	}
	imp, err := netchan.Import("tcp", masterAddr)
	if err != nil {
		log.Print("Ferrou")
	}

	c = make(chan WorkSummary, 10)
	imp.Import("masterChannel", c, netchan.Send, 10)
	go func() {
		err := <-imp.Errors()
		log.Print(err)
		log.Print("Recuperado")
	}()

	_sessions[session] = c
	return
}

func cacheWatcher() {
	for {
		time.Sleep(3000 * 1000000)
		mu.Lock()
		log.Print("Cleanning up Sessions")
		for k, _ := range _sessions {
			_sessions[k] = _sessions[k], false
		}
		mu.Unlock()
	}
}
//Listen to the worker channel. Every Task is executed by a different
//go routine 
func (w *LocalWorker) Serve() {
	log.Print("Waiting for tasks...")
	go cacheWatcher()
	for {
		task := <-w.channel
		if *w.mode == "worker" {
			w.SetMasterChan(importMasterChan(task.MasterAddr, task.Session))
		}

		log.Printf("Task Received from %v", task.MasterAddr)
		go w.execute(task)
	}
}

//Excecutes a task and send back a response to
//w.masterChannel. masterChannel can be set by 
//w.SetMasterChan in standalone mode or
//dinamically imported in worker mode        
func (w *LocalWorker) execute(task Task) {

	client := NewHTTPClient(task.Host, "")
	client.Auth(task.User, task.Password)
	var totalElapsed int64
	totalErr := 0
	totalSuc := 0
	var max int64 = 0
	var min int64 = -1
	//perform n times the request
	for i := 0; i < task.Requests; i++ {
		start := time.Nanoseconds()
		_, err := client.DoRequest()
		elapsed := (time.Nanoseconds() - start)
		if err == nil {
			totalSuc += 1
			totalElapsed += elapsed
			max = Max(max, totalElapsed/1000000)
			min = Min(min, totalElapsed/1000000)
		} else {
			totalErr += 1
		}
	}

	summary := &WorkSummary{
		ErrCount: totalErr,
		SucCount: totalSuc,
		Avg:      float64(totalElapsed / int64(totalSuc)),
		Max:      max,
		Min:      min,
	}

	w.masterChannel <- *summary
	log.Printf("Summary sent to %s", task.MasterAddr)
}
//Holds a reference to an imported channel
//from the actual worker
type ProxyWorker struct {
	channel  chan Task
	importer *netchan.Importer
}

//Creates a new Proxy importing 'workerChannel' from Worker running
//on workerAddr        
func NewProxyWorker(workerAddr string) (p *ProxyWorker, err os.Error) {
	log.Printf("Setting up a ProxyWorker for %s", workerAddr)
	p = new(ProxyWorker)

	imp, err := netchan.Import("tcp", workerAddr)
	if err != nil {
		return
	}
	p.importer = imp
	return
}

//Worker interface implemented
func (self *ProxyWorker) Channel() chan Task {
	return self.channel
}

func (self *ProxyWorker) Serve() {

	self.channel = make(chan Task)
	err := self.importer.Import("workerChannel", self.channel, netchan.Send, 10)
	if err != nil {
		log.Print(err)
	}
}
