package manager

import (
	"bufio"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
)

// helper function to log the stream of bytes from a reader while waiting on
// the error channel. It returns on first error received on the channel
func logOutputAndReturnStatus(r io.Reader, errCh chan error) error {
	// this can happen if an error occurred before the ansible could be run,
	// just return that error
	if r == nil {
		return <-errCh
	}

	s := bufio.NewScanner(r)
	ticker := time.Tick(50 * time.Millisecond)
	for {
		select {
		case err := <-errCh:
			for s.Scan() {
				log.Infof("%s", s.Bytes())
			}
			return err
		case <-ticker:
			// scan any available output while waiting
			if s.Scan() {
				log.Infof("%s", s.Bytes())
			}
		}
	}
}
