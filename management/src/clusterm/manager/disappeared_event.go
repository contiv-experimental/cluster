package manager

import (
	"fmt"

	"github.com/contiv/cluster/management/src/monitor"
)

// disappearedEvent processes the disappeared event from monitoring subsystem
type disappearedEvent struct {
	mgr   *Manager
	nodes []monitor.SubsysNode
}

// newDisappearedEvent creates and returns disappearedEvent event
func newDisappearedEvent(mgr *Manager, nodes []monitor.SubsysNode) *disappearedEvent {
	return &disappearedEvent{
		mgr:   mgr,
		nodes: nodes,
	}
}

func (e *disappearedEvent) String() string {
	return fmt.Sprintf("disappearedEvent: %+v", e.nodes[0])
}

func (e *disappearedEvent) process() error {
	//XXX: need to form the name that adheres to collins tag requirements
	name := e.nodes[0].GetLabel() + "-" + e.nodes[0].GetSerial()

	node, err := e.mgr.findNode(name)
	if err != nil {
		return err
	}

	// update node's monitoring info to the one received in the event.
	node.Mon = e.nodes[0]

	if err := e.mgr.inventory.SetAssetDisappeared(name); err != nil {
		// XXX. Log this to collins
		return err
	}
	return nil
}
