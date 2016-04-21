package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
)

// decommissionEvent triggers the decommission workflow
type decommissionEvent struct {
	mgr       *Manager
	nodeNames []string
	extraVars string
}

// newDecommissionEvent creates and returns decommissionEvent
func newDecommissionEvent(mgr *Manager, nodeNames []string, extraVars string) *decommissionEvent {
	return &decommissionEvent{
		mgr:       mgr,
		nodeNames: nodeNames,
		extraVars: extraVars,
	}
}

func (e *decommissionEvent) String() string {
	return fmt.Sprintf("decommissionEvent: %v", e.nodeNames)
}

func (e *decommissionEvent) process() error {
	if e.mgr.activeJob != nil {
		return errActiveJob(e.mgr.activeJob.String())
	}

	isMasterNode, err := e.mgr.isMasterNode(e.nodeNames[0])
	if err != nil {
		return err
	}

	// before setting the node cancelled and triggering the cleanup make sure
	// that the master node is decommissioned only if there are no more worker nodes.
	// XXX: revisit this check once we are able to support multiple master nodes.
	if isMasterNode {
		for name := range e.mgr.nodes {
			if name == e.nodeNames[0] {
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
				return errored.Errorf("%q is a master node and it can only be decommissioned after all worker nodes have been decommissioned", e.nodeNames[0])
			}
		}
	}

	if err := e.mgr.inventory.SetAssetCancelled(e.nodeNames[0]); err != nil {
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
			if err := e.mgr.inventory.SetAssetDecommissioned(e.nodeNames[0]); err != nil {
				// XXX. Log this to inventory
				log.Errorf("failed to update state in inventory, Error: %v", err)
			}
		})
	go e.mgr.activeJob.Run()

	return nil
}

// cleanupRunner is the job runner that runs cleanup playbooks on one or more nodes
func (e *decommissionEvent) cleanupRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	node, err := e.mgr.findNode(e.nodeNames[0])
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeNames[0])
	}

	outReader, cancelFunc, errCh := e.mgr.configuration.Cleanup(
		configuration.SubsysHosts([]*configuration.AnsibleHost{
			e.mgr.nodes[e.nodeNames[0]].Cfg.(*configuration.AnsibleHost),
		}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		return err
	}
	return nil
}
