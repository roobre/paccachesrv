package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"roob.re/paccachesrv"
	"roob.re/paccachesrv/cache"
)

const (
	envCache  = "PACSRV_CACHE"
	envMirror = "PACSRV_MIRROR"
)

func main() {
	path := os.Getenv(envCache)
	if path == "" {
		log.Fatalf("%s must be set", envCache)
	}

	mirror := os.Getenv(envMirror)
	if mirror == "" {
		log.Fatalf("%s must be set", envMirror)
	}

	cash := cache.New(path, mirror)
	srv := paccachesrv.New(cash)

	log.Infof("Starting server")
	err := http.ListenAndServe(":8000", srv)
	if err != nil {
		log.Error(err)
	}
}
