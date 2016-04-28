package manager

import (
	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/monitor"
)

func (m *Manager) enqueueMonitorEvent(events []monitor.Event) {
	// XXX: for now break the batch and inject one event per monitor event.
	// revisit later as batching requirements become more clear
	for _, e := range events {
		log.Debugf("processing monitor event: %+v", e)
		var me event
		switch e.Type {
		case monitor.Discovered:
			me = newDiscoveredEvent(m, e.Node)
		case monitor.Disappeared:
			me = newDisappearedEvent(m, e.Node)
		}
		m.reqQ <- me
		log.Debugf("enqueued manager event: %+v", me)
	}
}

func (m *Manager) monitorLoop(errCh chan error) {
	if err := m.monitor.Start(); err != nil {
		log.Errorf("monitoring subsystem encountered a failure. Error: %s", err)
		errCh <- err
	}
}
