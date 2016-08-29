package manager

import (
	"bufio"
	"io"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/errored"
)

var errJobCancelled = errored.Errorf("job was cancelled")

// helper function to log the stream of bytes from a reader while waiting on
// the error channel. It returns on first error received on the channel
func logOutputAndReturnStatus(r io.Reader, errCh chan error, cancelCh CancelChannel,
	cancelFunc context.CancelFunc, jobLogs io.Writer) error {
	// this can happen if an error occurred before the ansible could be run,
	// just return that error
	if r == nil {
		return <-errCh
	}

	// redirect read output to job logs
	t := io.TeeReader(r, jobLogs)
	s := bufio.NewScanner(t)
	ticker := time.Tick(50 * time.Millisecond)
	for {
		var err error
		select {
		case <-cancelCh:
			err = errJobCancelled
			cancelFunc()
			for s.Scan() {
				logrus.Infof("%s", s.Bytes())
			}
			return err
		case err := <-errCh:
			for s.Scan() {
				logrus.Infof("%s", s.Bytes())
			}
			return err
		case <-ticker:
			// scan any available output while waiting
			if s.Scan() {
				logrus.Infof("%s", s.Bytes())
			}
		}
	}
}

// commonEventValidate does common validation for events. It returns a map of nodes
// associted with their name on success
func (m *Manager) commonEventValidate(nodeNames []string) (map[string]*node, error) {
	if len(nodeNames) == 0 {
		return nil, errored.Errorf("atleast one node should be specified")
	}

	err := m.areDiscoveredNodes(nodeNames)
	if err != nil {
		return nil, err
	}

	enodes := map[string]*node{}
	for _, name := range nodeNames {
		node, err := m.findNode(name)
		if err != nil {
			return nil, err
		}
		if node.Cfg == nil {
			return nil, nodeConfigNotExistsError(name)
		}
		enodes[name] = node
	}

	return enodes, nil
}
