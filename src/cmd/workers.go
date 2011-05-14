package main

import (
	"log"
	"netchan"
	"time"
	"fmt"
)

//Represents a set of request to be performed
//against Task.Host        
type Task struct {
	Host, User, Password string
	BasicAuth            bool
	Requests, Id         int
	MasterAddr           string
}

//Put t to w.Channel()        
func (t *Task) Send(w Worker) {
	w.Channel() <- *t
}

//The worker interface
type Worker interface {
	//Should return the input channel to
        //interact with Worker
        Channel() chan Task
}


//A local workers is used in standalone mode
//as well as in worker mode.
type LocalWorker struct {
	//the Worker input channel to
        //receve tasks
        channel chan Task
}

//Creates a new LocalWorker. If export is true, than
//the LocalWorker exports its input channel in the network address
//provided by workerAddr        
func NewLocalWorker(export bool, workerAddr string) (w *LocalWorker) {
	log.Print("Setting up a Localworker...")
	w = new(LocalWorker)
	w.channel = make(chan Task)

	//exports the channels
	if export {
		e := netchan.NewExporter()
		e.Export("workerChannel", w.channel, netchan.Recv)
		e.ListenAndServe("tcp", workerAddr)
	}
	w.Execute()
	return
}

//Holds a reference to an imported channel
//from the actual worker
type ProxyWorker struct {
	channel chan Task
}

func (p *ProxyWorker) Channel() chan Task {
	return p.channel
}

func (l *LocalWorker) Channel() chan Task {
	return l.channel
}

//Creates a new Proxy importing 'workerChannel' from Worker running
//on workerAddr        
func NewProxyWorker(workerAddr string) (p *ProxyWorker) {
	log.Print("Setting up a ProxyWorker")
	p = new(ProxyWorker)
	imp, _ := netchan.Import("tcp", workerAddr)
	p.channel = make(chan Task)
	imp.ImportNValues("workerChannel", p.channel, netchan.Send, 1, 1)
	return p
}


//Helper function to import the Master channel from masterAddr
func importMasterChan(masterAddr string) (c chan WorkSummary) {
	imp, _ := netchan.Import("tcp", masterAddr)
	c = make(chan WorkSummary, 10)
	imp.Import("masterChannel", c, netchan.Send, 10)
	return
}

func (w *LocalWorker) Execute() {
	log.Print("Waiting for tasks...")

	for {
		task := <-w.channel
		log.Printf("Task Received from %v", task.MasterAddr)

		masterChannel := importMasterChan(task.MasterAddr)

		client := NewHTTPClient(task.Host, "")
		if task.BasicAuth {
			client.Auth(task.User, task.Password)
		}
		var totalElapsed int64
		totalErr := 0
		totalSuc := 0

		//perform n times the request
		for i := 0; i < task.Requests; i++ {
			start := time.Nanoseconds()
			_, err := client.DoRequest()
			elapsed := (time.Nanoseconds() - start)
			if err == nil {
				totalSuc += 1
				totalElapsed += elapsed
			} else {
				totalErr += 1
			}
		}

		summary := &WorkSummary{
			ErrCount: totalErr,
			SucCount: totalSuc,
			Avg:      float64(totalElapsed / int64(totalSuc)),
		}
		masterChannel <- *summary

		log.Print("Summary sent to %s", task.MasterAddr)
	}

}
//Reported by the worker through resultChan
type WorkSummary struct {
	ErrCount int     //total errors
	SucCount int     //total success
	Avg      float64 //average response time
}
