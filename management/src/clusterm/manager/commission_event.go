package manager

import (
	"fmt"
	"io"

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
	hostGroup string

	_hosts  configuration.SubsysHosts
	_enodes map[string]*node
}

// newCommissionEvent creates and returns commissionEvent
func newCommissionEvent(mgr *Manager, nodeNames []string, extraVars, hostGroup string) *commissionEvent {
	return &commissionEvent{
		mgr:       mgr,
		nodeNames: nodeNames,
		extraVars: extraVars,
		hostGroup: hostGroup,
	}
}

func (e *commissionEvent) String() string {
	return fmt.Sprintf("commissionEvent: nodes:%v extra-vars:%v host-group:%v",
		e.nodeNames, e.extraVars, e.hostGroup)
}

func (e *commissionEvent) process() error {
	// err shouldn't be redefined below
	var err error

	err = e.mgr.checkAndSetActiveJob(
		e.String(),
		e.configureOrCleanupOnErrorRunner,
		func(status JobStatus, errRet error) {
			if status == Errored {
				log.Errorf("configuration job failed. Error: %v", errRet)
				// set assets as unallocated
				e.mgr.setAssetsStatusBestEffort(e.nodeNames, e.mgr.inventory.SetAssetUnallocated)
				return
			}
			// set assets as commissioned
			e.mgr.setAssetsStatusBestEffort(e.nodeNames, e.mgr.inventory.SetAssetCommissioned)
		})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			e.mgr.resetActiveJob()
		}
	}()

	// validate event data
	if err = e.eventValidate(); err != nil {
		return err
	}

	// prepare inventory
	if err = e.prepareInventory(); err != nil {
		return err
	}

	// set assets as provisioning
	if err = e.mgr.setAssetsStatusAtomic(e.nodeNames, e.mgr.inventory.SetAssetProvisioning,
		e.mgr.inventory.SetAssetUnallocated); err != nil {
		return err
	}

	// trigger node configuration
	go e.mgr.runActiveJob()

	return nil
}

func (e *commissionEvent) eventValidate() error {
	var err error
	e._enodes, err = e.mgr.commonEventValidate(e.nodeNames)
	if err != nil {
		return err
	}

	if !IsValidHostGroup(e.hostGroup) {
		return errored.Errorf("invalid or empty host-group specified: %q", e.hostGroup)
	}

	// when workers are being configured, make sure that there is atleast one service-master
	if e.hostGroup == ansibleWorkerGroupName {
		masterCommissioned := false
		for name := range e.mgr.nodes {
			if _, ok := e._enodes[name]; ok {
				// skip nodes in the event
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

			// found a master node
			masterCommissioned = true
			break
		}
		if !masterCommissioned {
			return errored.Errorf("Cannot commission a worker node without existence of a master node in the cluster, make sure atleast one master node is commissioned.")
		}
	}
	return nil
}

// prepareInventory adds the specifed nodes to the specified host-group
func (e *commissionEvent) prepareInventory() error {
	hosts := []*configuration.AnsibleHost{}
	for _, node := range e._enodes {
		hostInfo := node.Cfg.(*configuration.AnsibleHost)
		hostInfo.SetGroup(e.hostGroup)
		hosts = append(hosts, hostInfo)
	}
	e._hosts = hosts

	return nil
}

// configureOrCleanupOnErrorRunner is the job runner that runs configuration playbooks on one or more nodes.
// It runs cleanup playbook on failure
func (e *commissionEvent) configureOrCleanupOnErrorRunner(cancelCh CancelChannel, jobLogs io.Writer) error {
	outReader, cancelFunc, errCh := e.mgr.configuration.Configure(e._hosts, e.extraVars)
	cfgErr := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs)
	if cfgErr == nil {
		return nil
	}
	log.Errorf("configuration failed, starting cleanup. Error: %s", cfgErr)
	outReader, cancelFunc, errCh = e.mgr.configuration.Cleanup(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs); err != nil {
		log.Errorf("cleanup failed. Error: %s", err)
	}

	//return the error status from provisioning
	return cfgErr
}
