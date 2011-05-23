// Copyright (c) Paulo Suzart. All rights reserved.
// The use and distribution terms for this software are covered by the
// Eclipse Public License 1.0 (http://opensource.org/licenses/eclipse-1.0.php)
// which can be found in the file epl-v10.html at the root of this distribution.
// By using this software in any fashion, you are agreeing to be bound by
// the terms of this license.
// You must not remove this notice, or any other, from this software.

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

var host, _ = os.Hostname()
var (
	mode     = flag.String("M", "standalone", "standalone, master, worker")
	maxTime  = flag.Int64("T", -1, "Max time in milisecs.")
	hostAddr = flag.String("H", host+":1970", "The master Addr")
)

func init() {
	flag.Parse()
	log.Printf("Starting in %s mode - %s", *mode, *hostAddr)
}

func main() {
	switch *mode {
	case "master", "standalone":
		m := NewMaster(mode, hostAddr, *maxTime*1000000)
		ctrlChan := make(chan bool)
		m.BenchMark(ctrlChan)
		if *maxTime != -1 {
			go supervise(m, maxTime)
		}
		<-ctrlChan
		log.Print(m.summary)
	case "worker":
		NewLocalWorker(mode, hostAddr).Serve()
	}
}

func supervise(supervised Supervised, maxTime *int64) {
	time.Sleep(*maxTime * 1000000)
	log.Print("WARN! gb stopped due to timeout. Work lost.")
	supervised.Shutdown()
	os.Exit(1)
}
