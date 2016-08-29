package manager

import "github.com/Sirupsen/logrus"

// event associates an event to corresponding processing logic
type event interface {
	String() string
	process() error
}

func (m *Manager) eventLoop() {
	for {
		me := <-m.reqQ
		logrus.Debugf("dequeued manager event: %+v", me)
		if err := me.process(); err != nil {
			// log and continue
			logrus.Errorf("error handling event %q. Error: %s", me, err)
		}
	}
}
