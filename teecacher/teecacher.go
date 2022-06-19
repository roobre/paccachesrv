package teecacher

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type teeCacher struct {
	reader   io.Reader
	closer   func() error
	filepath string
	file     io.WriteCloser
	readErr  error
}

func TeeCacher(reader io.Reader, filepath string) (io.ReadCloser, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("creating file %q: %w", filepath, err)
	}

	closer := func() error { return nil }
	if rc, hasClose := reader.(io.ReadCloser); hasClose {
		closer = rc.Close
	}

	return &teeCacher{
		reader:   io.TeeReader(reader, file),
		closer:   closer,
		filepath: filepath,
		file:     file,
	}, nil
}

func (tc *teeCacher) Read(p []byte) (int, error) {
	var n int
	n, tc.readErr = tc.reader.Read(p)
	return n, tc.readErr
}

func (tc *teeCacher) Close() error {
	var closeErr error
	if tc.closer != nil {
		closeErr = tc.closer()
	}

	if tc.readErr != nil && !errors.Is(tc.readErr, io.EOF) {
		log.Warnf("Error occurred while reading reader, deleting %q: %v", tc.filepath, tc.readErr)
		err := os.Remove(tc.filepath)

		if err != nil {
			return fmt.Errorf("removing cache file: %v", err)
		}

		return closeErr
	}

	return closeErr
}
