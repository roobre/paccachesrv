package paccachesrv_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"roob.re/paccachesrv"
	"roob.re/paccachesrv/cache"
	"strings"
	"sync"
	"testing"
)

const (
	existentContents = "I am an existing file!"
	newContents      = "I am a cache miss"
)

func TestServer(t *testing.T) {
	tmp := t.TempDir()

	file, err := os.Create(path.Join(tmp, "existent"))
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(file, strings.NewReader(existentContents))

	callsMtx := &sync.Mutex{}
	calls := map[string]int{}
	mirrorMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		callsMtx.Lock()
		defer callsMtx.Unlock()

		calls[r.URL.Path] = calls[r.URL.Path] + 1
		rw.Write([]byte(newContents))
	}))

	t.Cleanup(func() {
		mirrorMock.Close()
	})

	server := paccachesrv.New(cache.New(tmp, mirrorMock.URL))
	testServer := httptest.NewServer(server)
	t.Cleanup(func() {
		testServer.Close()
	})

	testUrl := testServer.URL
	testClient := testServer.Client()

	t.Run("hits_cache", func(t *testing.T) {
		resp, err := testClient.Get(testUrl + "/existent")
		if err != nil {
			t.Fatal(err)
		}

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(contents, []byte(existentContents)) {
			t.Fatal("contents are not equal")
		}

		if calls["/existent"] != 0 {
			t.Fatal("mirror mock was called for an expected cache hit")
		}
	})

	t.Run("misses_cache", func(t *testing.T) {
		resp, err := testClient.Get(testUrl + "/new")
		if err != nil {
			t.Fatal(err)
		}

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(contents, []byte(newContents)) {
			t.Fatal("contents are not equal")
		}

		if calls["/new"] != 1 {
			t.Fatal("mirror mock was not called for an expected cache miss")
		}
	})

	t.Run("hits_subsequently", func(t *testing.T) {
		resp, err := testClient.Get(testUrl + "/new")
		if err != nil {
			t.Fatal(err)
		}

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(contents, []byte(newContents)) {
			t.Fatal("contents are not equal")
		}

		if calls["/new"] != 1 {
			t.Fatal("mirror mock was called for an expected cache hit")
		}
	})
}
