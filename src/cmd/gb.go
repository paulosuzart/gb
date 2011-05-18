package main

import (
	"flag"
	"log"
	"time"
	"os"
)

type Supervised interface {
	Shutdown()
}

var (
	mode     = flag.String("M", "standalone", "standalone, master, worker")
	maxTime  = flag.Int64("T", -1, "Max time in milisecs.")
	hostAddr = flag.String("H", "localhost:1970", "The master Addr")
)

func main() {
	flag.Parse()
	log.Printf("Starting in %s mode", *mode)
	switch *mode {
	case "master", "standalone":
		m := NewMaster(mode, hostAddr)
		m.BenchMark()
		if *maxTime != -1 {
			go supervise(m, maxTime)
		}
		<-m.ctrlChan
	case "worker":
		w := NewLocalWorker(mode, hostAddr)
		if *maxTime != -1 {
			go supervise(nil, maxTime)
		}
		<-w.ctrlChan
	}
}

func supervise(supervised Supervised, maxTime *int64) {
	time.Sleep(*maxTime * 1000000)
	log.Print("WARN! gb stopped due to timeout. Work lost.")
        supervised.Shutdown()        
	os.Exit(1)
}
