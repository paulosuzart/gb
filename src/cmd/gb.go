package main

import (
	"flag"
	"log"
)


var (
	mode = flag.String("M", "standalone", "standalone, master, worker")

	hostAddr = flag.String("H", "localhost:9393", "The master Addr")
)

func main() {
	flag.Parse()
	log.Printf("Starting in %s mode", *mode)
	if *mode == "master" || *mode == "standalone" {
		m := NewMaster(mode, hostAddr)
		m.BenchMark()
		<-m.ctrlChan
	} else if *mode == "worker" {
		w := NewLocalWorker(mode, hostAddr)
		<-w.ctrlChan
	}
}
