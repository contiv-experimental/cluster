package manager

import (
	"fmt"

	"github.com/contiv/errored"
)

// nodeCommissioned triggers the commission workflow
type nodeCommissioned struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

// newNodeCommissioned creates and returns nodeCommissioned event
func newNodeCommissioned(mgr *Manager, nodeName, extraVars string) *nodeCommissioned {
	return &nodeCommissioned{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeCommissioned) String() string {
	return fmt.Sprintf("nodeCommissioned: %s", e.nodeName)
}

func (e *nodeCommissioned) process() error {
	isDiscovered, err := e.mgr.isDiscoveredNode(e.nodeName)
	if err != nil {
		return err
	}
	if !isDiscovered {
		return errored.Errorf("node %q has disappeared from monitoring subsystem, it can't be commissioned. Please check node's network reachability", e.nodeName)
	}

	if err := e.mgr.inventory.SetAssetProvisioning(e.nodeName); err != nil {
		// XXX. Log this to collins
		return err
	}

	// trigger node configuration event
	e.mgr.reqQ <- newNodeConfigure(e.mgr, e.nodeName, e.extraVars)
	return nil
}
