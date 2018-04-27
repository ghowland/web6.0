package main

import (
	"github.com/ghowland/web6.0/web6"
	//go RunJobWorkers()


	"github.com/jcasts/gosrv"
)

func main() {
	pidFile := "" // Pid file passed from flag (for prod)

	web6.Start(&pidFile)

	//TODO(g): Process all the gosrv flags ourselves if necessary (in web6.Start()) and set the values before starting up (make sure pid is properly assigned)
	s := gosrv.New()

	// Update Server settings from flags if necessary
	if pidFile != "" && pidFile != gosrv.DefaultPidFile {
		s.PidFile = pidFile
	}

	// We dont use NewFromFlag() because it doesnt allow adding custom flags.  We have to manage the gosrv config Struct ourselves in config.go
	/*
	s, err := gosrv.NewFromFlag()
	if err != nil {
		panic("Cannot create web server: " + err.Error() + "\n")
	}*/

	s.HandleFunc("/", web6.Handler)

	err := s.ListenAndServe()
	if err != nil {
		panic("Cannot listen as web server: " + err.Error() + "\n")
	}
}
