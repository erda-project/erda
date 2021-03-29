package common

import (
	"compress/gzip"
	"io"
)

// CompressWithGzip takes an io.Reader as input and pipes
// it through a gzip.Writer returning an io.Reader containing
// the gzipped data.
// An error is returned if passing data to the gzip.Writer fails
func CompressWithGzip(data io.Reader) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)

	var err error
	go func() {
		_, err = io.Copy(gzipWriter, data)
		gzipWriter.Close()
		// subsequent reads from the read half of the pipe will
		// return no bytes and the error err, or EOF if err is nil.
		pipeWriter.CloseWithError(err)
	}()

	return pipeReader, err
}
