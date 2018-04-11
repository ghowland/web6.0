package main

import (
	"github.com/ghowland/web6.0/web6"
	//go RunJobWorkers()


	"github.com/jcasts/gosrv"
)

func main() {
	web6.Start()

	//TODO(g): Process all the gosrv flags ourselves (in web6.Start()) and set the values before starting up, then remove below commented code
	s := gosrv.New()
	/*
	s, err := gosrv.NewFromFlag()
	if err != nil {
		panic("Cannot create web server: " + err.Error() + "\n")
	}
	*/

	s.HandleFunc("/", web6.Handler)

	err := s.ListenAndServe()
	if err != nil {
		panic("Cannot listen as web server: " + err.Error() + "\n")
	}
}
