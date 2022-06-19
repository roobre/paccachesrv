package cache

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"roob.re/paccachesrv/teecacher"
	"strings"
)

type HTTPError int

func (he HTTPError) Error() string {
	return fmt.Sprintf("received non-200 code %d", int(he))
}

type Cache struct {
	storagePath string
	client      *http.Client
	mirror      string
}

func New(path, mirror string) *Cache {
	return &Cache{
		storagePath: path,
		mirror:      mirror,
		client:      http.DefaultClient,
	}
}

func (c *Cache) pathFor(requestPath string) string {
	return path.Join(c.storagePath, path.Base(requestPath))
}

func (c *Cache) openRead(requestPath string) (io.ReadCloser, error) {
	return os.Open(c.pathFor(requestPath))
}

func (c *Cache) url(requestPath string) string {
	return strings.TrimSuffix(c.mirror, "/") + "/" + strings.TrimPrefix(requestPath, "/")
}

func (c *Cache) shouldCache(requestPath string) bool {
	return strings.Contains(path.Base(requestPath), ".pkg.")
}

func (c *Cache) Reader(requestPath string) (io.ReadCloser, error) {
	if file, err := c.openRead(requestPath); err == nil {
		log.Infof("cache hit for %s", requestPath)
		return file, nil
	}

	log.Infof("%s not found in cache", requestPath)

	resp, err := c.client.Get(c.url(requestPath))
	if err != nil {
		return nil, fmt.Errorf("upstream mirror returned an error: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, HTTPError(resp.StatusCode)
	}

	if c.shouldCache(requestPath) {
		return teecacher.TeeCacher(resp.Body, c.pathFor(requestPath))
	}

	return resp.Body, nil
}
