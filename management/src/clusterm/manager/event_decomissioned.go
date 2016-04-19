package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
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
	if e.mgr.activeJob != nil {
		return errActiveJob(e.mgr.activeJob.String())
	}

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
		// XXX. Log this to inventory
		return err
	}
	// trigger node clenup
	e.mgr.activeJob = NewJob(
		e.cleanupRunner,
		func(status JobStatus, errRet error) {
			if status == Errored {
				log.Errorf("cleanup job failed. Error: %v", errRet)
			}

			// set asset state to decommissioned
			if err := e.mgr.inventory.SetAssetDecommissioned(e.nodeName); err != nil {
				// XXX. Log this to inventory
				log.Errorf("failed to update state in inventory, Error: %v", err)
			}
		})
	go e.mgr.activeJob.Run()

	return nil
}

// cleanupRunner is the job runner that runs cleanup playbooks on one or more nodes
func (e *nodeDecommissioned) cleanupRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	node, err := e.mgr.findNode(e.nodeName)
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeName)
	}

	outReader, cancelFunc, errCh := e.mgr.configuration.Cleanup(
		configuration.SubsysHosts([]*configuration.AnsibleHost{
			e.mgr.nodes[e.nodeName].Cfg.(*configuration.AnsibleHost),
		}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		return err
	}
	return nil
}
