package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
)

func errActiveJob(desc string) error {
	return errored.Errorf("there is already an active job, please try in sometime. Job: %s", desc)
}

// commissionEvent triggers the commission workflow
type commissionEvent struct {
	mgr       *Manager
	nodeNames []string
	extraVars string
}

// newCommissionEvent creates and returns commissionEvent
func newCommissionEvent(mgr *Manager, nodeNames []string, extraVars string) *commissionEvent {
	return &commissionEvent{
		mgr:       mgr,
		nodeNames: nodeNames,
		extraVars: extraVars,
	}
}

func (e *commissionEvent) String() string {
	return fmt.Sprintf("commissionEvent: %v", e.nodeNames)
}

func (e *commissionEvent) process() error {
	if e.mgr.activeJob != nil {
		return errActiveJob(e.mgr.activeJob.String())
	}

	isDiscovered, err := e.mgr.isDiscoveredNode(e.nodeNames[0])
	if err != nil {
		return err
	}
	if !isDiscovered {
		return errored.Errorf("node %q has disappeared from monitoring subsystem, it can't be commissioned. Please check node's network reachability", e.nodeNames[0])
	}

	if err := e.mgr.inventory.SetAssetProvisioning(e.nodeNames[0]); err != nil {
		// XXX. Log this to collins
		return err
	}

	// trigger node configuration
	e.mgr.activeJob = NewJob(
		e.configureOrCleanupOnErrorRunner,
		func(status JobStatus, errRet error) {
			if status == Errored {
				log.Errorf("configuration job failed. Error: %v", errRet)
				// set asset state back to unallocated
				if err := e.mgr.inventory.SetAssetUnallocated(e.nodeNames[0]); err != nil {
					// XXX. Log this to inventory
					log.Errorf("failed to update state in inventory, Error: %v", err)
				}
				return
			}
			// set asset state to commissioned
			if err := e.mgr.inventory.SetAssetCommissioned(e.nodeNames[0]); err != nil {
				// XXX. Log this to inventory
				log.Errorf("failed to update state in inventory, Error: %v", err)
			}
		})
	go e.mgr.activeJob.Run()

	return nil
}

// configureOrCleanupOnErrorRunner is the job runner that runs configuration playbooks on one or more nodes.
// It runs cleanup playbook on failure
func (e *commissionEvent) configureOrCleanupOnErrorRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	node, err := e.mgr.findNode(e.nodeNames[0])
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeNames[0])
	}

	hostInfo := node.Cfg.(*configuration.AnsibleHost)
	nodeGroup := ansibleMasterGroupName
	masterAddr := ""
	masterName := ""
	// update the online master address if this is second node that is being commissioned.
	// Also set the group for second or later nodes to be worker, as right now services like
	// swarm and netmaster can only have one master node and also we don't yet have a vip
	// service.
	// XXX: revisit this when the above changes
	for name, node := range e.mgr.nodes {
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

		isMasterNode, err := e.mgr.isMasterNode(name)
		if err != nil || !isMasterNode {
			if err != nil {
				log.Debugf("a node check failed for %q. Error: %s", name, err)
			}
			//skip the hosts that are not in master group
			continue
		}

		// found our node
		masterAddr = node.Mon.GetMgmtAddress()
		masterName = node.Cfg.GetTag()
		nodeGroup = ansibleWorkerGroupName
		break
	}
	hostInfo.SetGroup(nodeGroup)
	hostInfo.SetVar(ansibleEtcdMasterAddrHostVar, masterAddr)
	hostInfo.SetVar(ansibleEtcdMasterNameHostVar, masterName)

	outReader, cancelFunc, errCh := e.mgr.configuration.Configure(
		configuration.SubsysHosts([]*configuration.AnsibleHost{hostInfo}), e.extraVars)
	cfgErr := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc)
	if cfgErr == nil {
		return nil
	}
	log.Errorf("configuration failed. Error: %s", cfgErr)
	outReader, cancelFunc, errCh = e.mgr.configuration.Cleanup(
		configuration.SubsysHosts([]*configuration.AnsibleHost{hostInfo}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		log.Errorf("cleanup failed. Error: %s", err)
	}

	//return the error status from provisioning
	return cfgErr
}
