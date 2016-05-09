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
		eventName := ""
		switch e.Type {
		case monitor.Discovered:
			eventName = monitor.Discovered.String()
		case monitor.Disappeared:
			eventName = monitor.Disappeared.String()
		default:
			log.Errorf("unexpected monitor event type %v", e.Type)
			continue
		}
		if err := NewClient(m.addr).PostMonitorEvent(eventName,
			[]MonitorNode{
				{
					Label:    e.Node.GetLabel(),
					Serial:   e.Node.GetSerial(),
					MgmtAddr: e.Node.GetMgmtAddress(),
				},
			}); err != nil {
			log.Errorf("error posting monitor event %q. Error: %v", eventName, err)
		}
	}
}

func (m *Manager) monitorLoop(errCh chan error) {
	if err := m.monitor.Start(); err != nil {
		log.Errorf("monitoring subsystem encountered a failure. Error: %s", err)
		errCh <- err
	}
}
