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

	_hosts  configuration.SubsysHosts
	_enodes map[string]*node
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

	// validate event data
	var err error
	if e._enodes, err = e.mgr.commonEventValidate(e.nodeNames); err != nil {
		return err
	}

	// prepare inventory
	if err := e.prepareInventory(); err != nil {
		return err
	}

	// set assets as cancelled
	if err := e.mgr.setAssetsStatusAtomic(e.nodeNames, e.mgr.inventory.SetAssetCancelled,
		e.mgr.inventory.SetAssetCommissioned); err != nil {
		return err
	}

	// trigger node cleanup
	e.mgr.activeJob = NewJob(
		e.cleanupRunner,
		func(status JobStatus, errRet error) {
			if status == Errored {
				log.Errorf("cleanup job failed. Error: %v", errRet)
			}

			// set assets as decommissioned
			e.mgr.setAssetsStatusBestEffort(e.nodeNames, e.mgr.inventory.SetAssetDecommissioned)
		})
	go e.mgr.activeJob.Run()

	return nil
}

// prepareInventory validates that after the cleanup on the nodes in the event,
// one of following is still true:
// - all nodes have been cleaned up; OR
// - there is atleast one master node left
func (e *decommissionEvent) prepareInventory() error {
	mastersLeft := 0
	workersLeft := 0
	for name := range e.mgr.nodes {
		if _, ok := e._enodes[name]; ok {
			// skip the node in the event
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
			workersLeft++
		} else {
			mastersLeft++
		}
	}

	if workersLeft > 0 && mastersLeft <= 0 {
		return errored.Errorf("decommissioning the specified node(s) will leave only worker nodes in the cluster, make sure all worker nodes are decommissioned before last master node.")
	}

	// prepare the inventory
	hosts := []*configuration.AnsibleHost{}
	for _, node := range e._enodes {
		hosts = append(hosts, node.Cfg.(*configuration.AnsibleHost))
	}
	e._hosts = hosts

	return nil
}

// cleanupRunner is the job runner that runs cleanup playbooks on one or more nodes
func (e *decommissionEvent) cleanupRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	outReader, cancelFunc, errCh := e.mgr.configuration.Cleanup(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		return err
	}
	return nil
}
