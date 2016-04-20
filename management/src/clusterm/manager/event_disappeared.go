package manager

import (
	"fmt"

	"github.com/contiv/cluster/management/src/monitor"
)

// nodeDisappeared processes the disappeared event from monitoring subsytem
type nodeDisappeared struct {
	mgr  *Manager
	node monitor.SubsysNode
}

// newNodeDisappeared creates and returns nodeDisappeared event
func newNodeDisappeared(mgr *Manager, node monitor.SubsysNode) *nodeDisappeared {
	return &nodeDisappeared{
		mgr:  mgr,
		node: node,
	}
}

func (e *nodeDisappeared) String() string {
	return fmt.Sprintf("nodeDisappeared: %+v", e.node)
}

func (e *nodeDisappeared) process() error {
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
