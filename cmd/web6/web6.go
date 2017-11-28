package main

import (
	"github.com/ghowland/web6.0/web6"
	"github.com/ghowland/yudien/yudien"

	"github.com/jcasts/gosrv"
)

func main() {
	////DEBUG: Testing
	//TestUdn()

	//go RunJobWorkers()
	web6.LoadConfig()
	yudien.Configure(&web6.Config.Ldap, &web6.Config.Opsdb)

	s, err := gosrv.NewFromFlag()
	if err != nil {
		panic("Cannot create web server: " + err.Error() + "\n")
	}

	s.HandleFunc("/", web6.Handler)

	err = s.ListenAndServe()
	if err != nil {
		panic("Cannot listen as web server: " + err.Error() + "\n")
	}
}
