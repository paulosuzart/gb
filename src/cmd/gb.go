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
	//standalone mode not supported by now.
	if *mode == "standalone" {
		log.Panic("Standalone mode not supported yet.")
	}
	log.Printf("Starting in %s mode", *mode)
	if *mode == "master" || *mode == "standalone" {
		m := NewMaster(mode, hostAddr)
		m.BenchMark()
		<-m.ctrlChan
	} else if *mode == "worker" {
		NewLocalWorker(mode, hostAddr)
	}
}
