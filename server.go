package paccachesrv

import (
	log "github.com/sirupsen/logrus"

	"io"
	"net/http"
	"roob.re/paccachesrv/cache"
)

type Server struct {
	cache *cache.Cache
}

func New(c *cache.Cache) *Server {
	return &Server{
		cache: c,
	}
}

func (s Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	reader, err := s.cache.Reader(r.URL.Path)
	if httpErr, isHttp := err.(cache.HTTPError); isHttp {
		rw.WriteHeader(int(httpErr))
		return
	} else if err != nil {
		log.Errorf("Creating reader: %v", err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	defer func() {
		err := reader.Close()
		if err != nil {
			log.Errorf("Closing cached reader: %v", err)
		}
	}()

	_, err = io.Copy(rw, reader)
	if err != nil {
		log.Errorf("Reading from cached reader: %v", err)
	}

	return
}
