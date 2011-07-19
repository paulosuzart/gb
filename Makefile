include $(GOROOT)/src/Make.inc
TARG=gb
GOFILES=util.go\
        http.go\
        workers.go\
        master.go\
        gb.go\
	

include $(GOROOT)/src/Make.cmd

