package main

import (
	"log"
	"netchan"
	"time"
	"os"
)

//Represents a set of request to be performed
//against Task.Host        
type Task struct {
	Host, User, Password string
	BasicAuth            bool
	Requests, Id         int
	MasterAddr           string
}

//Reported by the worker through resultChan
type WorkSummary struct {
	ErrCount int     //total errors
	SucCount int     //total success
	Avg      float64 //average response time
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

func (p *ProxyWorker) Channel() chan Task {
	return p.channel
}

func (l *LocalWorker) Channel() chan Task {
	return l.channel
}


//Creates a new LocalWorker. If export is true, than
//the LocalWorker exports its input channel in the network address
//provided by workerAddr        
func NewLocalWorker(mode, hostAddr *string) (w *LocalWorker) {
	log.Print("Setting up a Localworker...")
	w = new(LocalWorker)
	w.channel = make(chan Task, 10)

	//exports the channels
	if *mode == "worker" {
		e := netchan.NewExporter()
		e.Export("workerChannel", w.channel, netchan.Recv)
		e.ListenAndServe("tcp", *hostAddr)
	}
	w.start()
	return
}

//Holds a reference to an imported channel
//from the actual worker
type ProxyWorker struct {
	channel chan Task
}

//Creates a new Proxy importing 'workerChannel' from Worker running
//on workerAddr        
func NewProxyWorker(workerAddr string) (p *ProxyWorker, err os.Error) {
	log.Print("Setting up a ProxyWorker")
	p = new(ProxyWorker)
	imp, err := netchan.Import("tcp", workerAddr)
	if err != nil {
		return
	}
	p.channel = make(chan Task)
	err = imp.Import("workerChannel", p.channel, netchan.Send, 10)
	if err != nil {
		return
	}
	return
}

//Helper function to import the Master channel from masterAddr
func importMasterChan(masterAddr string) (c chan WorkSummary) {
       	imp, _ := netchan.Import("tcp", masterAddr)
	c = make(chan WorkSummary, 10)
	imp.Import("masterChannel", c, netchan.Send, 10)
        go func(){
           err := <- imp.Errors()
           log.Print(err)
        }()        
	return
}
func (w *LocalWorker) start() {
	log.Print("Waiting for tasks...")
	for {
		task := <-w.channel
		log.Printf("Task Received from %v", task.MasterAddr)
		go w.execute(task, importMasterChan(task.MasterAddr))
	}
}


func (w *LocalWorker) execute(task Task, masterChannel chan WorkSummary) {

	client := NewHTTPClient(task.Host, "")
	client.Auth(task.User, task.Password)
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
	log.Printf("Summary sent to %s", task.MasterAddr)
}
