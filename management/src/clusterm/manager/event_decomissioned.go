package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/errored"
)

// nodeDecommissioned triggers the decommission workflow
type nodeDecommissioned struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

// newNodeDecommissioned creates and returns nodeDecommissioned event
func newNodeDecommissioned(mgr *Manager, nodeName, extraVars string) *nodeDecommissioned {
	return &nodeDecommissioned{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeDecommissioned) String() string {
	return fmt.Sprintf("nodeDecommissioned: %s", e.nodeName)
}

func (e *nodeDecommissioned) process() error {
	isMasterNode, err := e.mgr.isMasterNode(e.nodeName)
	if err != nil {
		return err
	}

	// before setting the node cancelled and triggering the cleanup make sure
	// that the master node is decommissioned only if there are no more worker nodes.
	// XXX: revisit this check once we are able to support multiple master nodes.
	if isMasterNode {
		for name := range e.mgr.nodes {
			if name == e.nodeName {
				// skip this node
				continue
			}

			isDiscoveredAndAllocated, err := e.mgr.isDiscoveredAndAllocatedNode(name)
			if err != nil || !isDiscoveredAndAllocated {
				if err != nil {
					log.Debugf("a node check failed for %q. Error: %s", name, err)
				}
				// skip hosts that are not yet provisioned or not in discovered state
				continue
			}

			isWorkerNode, err := e.mgr.isWorkerNode(name)
			if err != nil {
				// skip this node
				log.Debugf("a node check failed for %q. Error: %s", name, err)
				continue
			}

			if isWorkerNode {
				return errored.Errorf("%q is a master node and it can only be decommissioned after all worker nodes have been decommissioned", e.nodeName)
			}
		}
	}

	if err := e.mgr.inventory.SetAssetCancelled(e.nodeName); err != nil {
		// XXX. Log this to collins
		return err
	}
	// trigger node clenup event
	e.mgr.reqQ <- newNodeCleanup(e.mgr, e.nodeName, e.extraVars)
	return nil
}
