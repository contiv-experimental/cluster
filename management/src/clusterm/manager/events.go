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
		logrus.Debugf("dequeued manager event: %s", me)
		err := me.process()
		// log and continue
		logrus.Debugf("done handling event %s. Error(if any): %v", me, err)
	}
}
