package main

import (
	"web6.0/web6"
	"github.com/jcasts/gosrv"
)


func main() {
	////DEBUG: Testing
	//TestUdn()

	//go RunJobWorkers()

	s, err := gosrv.NewFromFlag()
	if err != nil {
		panic(err)
	}

	s.HandleFunc("/", web6.Handler)

	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

