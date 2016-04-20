package manager

import "fmt"

// nodeInMaintenance triggers the upgrade workflow
type nodeInMaintenance struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

// newNodeInMaintenance creates and returns nodeInMaintenance event
func newNodeInMaintenance(mgr *Manager, nodeName, extraVars string) *nodeInMaintenance {
	return &nodeInMaintenance{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeInMaintenance) String() string {
	return fmt.Sprintf("nodeInMaintenance: %s", e.nodeName)
}

func (e *nodeInMaintenance) process() error {
	if err := e.mgr.inventory.SetAssetInMaintenance(e.nodeName); err != nil {
		// XXX. Log this to collins
		return err
	}
	// trigger node upgrade event
	e.mgr.reqQ <- newNodeUpgrade(e.mgr, e.nodeName, e.extraVars)
	return nil
}
