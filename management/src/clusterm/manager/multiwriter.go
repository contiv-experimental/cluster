package manager

import (
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
)

// MultiWriter allows writing to multiple writers. It is different from
// io.MultiWriter() in that it allows adding writers on the fly. The writers
// added later will only see data from the point of addition.
type MultiWriter struct {
	once    sync.Once
	writers map[io.Writer]struct{}
}

// Write writes to all the underlying writers. If write to a writer fails
// then it is evicted from the map
func (mw *MultiWriter) Write(p []byte) (int, error) {
	for w := range mw.writers {
		if _, err := w.Write(p); err != nil {
			logrus.Debugf("failed to write to writer %+v", w)
			delete(mw.writers, w)
		}
	}
	return len(p), nil
}

// Close closes the underlying writers if they implement WriteCloser
func (mw *MultiWriter) Close() error {
	for w := range mw.writers {
		if wc, ok := w.(io.WriteCloser); ok {
			wc.Close()
		}
	}
	return nil
}

// Add adds a writer to the list of writers
func (mw *MultiWriter) Add(w io.Writer) {
	mw.once.Do(func() { mw.writers = make(map[io.Writer]struct{}) })
	mw.writers[w] = struct{}{}
}
