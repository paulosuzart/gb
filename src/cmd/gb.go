package main

import (
	"flag"
	"log"
)


var (
	mode = flag.String("M", "standalone", "standalone, master, worker")

	hostAddr = flag.String("H", "localhost:1970", "The master Addr")
)

func main() {
	flag.Parse()
	log.Printf("Starting in %s mode", *mode)
	switch *mode {
	case "master", "standalone":
		m := NewMaster(mode, hostAddr)
		m.BenchMark()
		<-m.ctrlChan
	case "worker":
		w := NewLocalWorker(mode, hostAddr)
		<-w.ctrlChan
	}
}
