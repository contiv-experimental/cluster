package manager

import (
	"fmt"

	"github.com/contiv/cluster/management/src/monitor"
)

// disappearedEvent processes the disappeared event from monitoring subsytem
type disappearedEvent struct {
	mgr  *Manager
	node monitor.SubsysNode
}

// newDisappearedEvent creates and returns disappearedEvent event
func newDisappearedEvent(mgr *Manager, node monitor.SubsysNode) *disappearedEvent {
	return &disappearedEvent{
		mgr:  mgr,
		node: node,
	}
}

func (e *disappearedEvent) String() string {
	return fmt.Sprintf("disappearedEvent: %+v", e.node)
}

func (e *disappearedEvent) process() error {
	//XXX: need to form the name that adheres to collins tag requirements
	name := e.node.GetLabel() + "-" + e.node.GetSerial()

	node, err := e.mgr.findNode(name)
	if err != nil {
		return err
	}

	// update node's monitoring info to the one received in the event.
	node.Mon = e.node

	if err := e.mgr.inventory.SetAssetDisappeared(name); err != nil {
		// XXX. Log this to collins
		return err
	}
	return nil
}
