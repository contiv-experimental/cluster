package manager

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
)

// updateEvent triggers the upgrade workflow
type updateEvent struct {
	mgr       *Manager
	nodeNames []string
	extraVars string
	hostGroup string

	_hosts  configuration.SubsysHosts
	_enodes map[string]*node
}

// newUpdateEvent creates and returns updateEvent
func newUpdateEvent(mgr *Manager, nodeNames []string, extraVars, hostGroup string) *updateEvent {
	return &updateEvent{
		mgr:       mgr,
		nodeNames: nodeNames,
		extraVars: extraVars,
		hostGroup: hostGroup,
	}
}

func (e *updateEvent) String() string {
	return fmt.Sprintf("updateEvent: nodes: %v extra-vars: %v host-group: %q", e.nodeNames, e.extraVars, e.hostGroup)
}

func (e *updateEvent) process() error {
	// err shouldn't be redefined below
	var err error

	err = e.mgr.checkAndSetActiveJob(
		e.String(),
		e.updateRunner,
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
	if err = e.pepareInventory(); err != nil {
		return err
	}

	//set assets as in-maintenance
	if err = e.mgr.setAssetsStatusAtomic(e.nodeNames, e.mgr.inventory.SetAssetInMaintenance,
		e.mgr.inventory.SetAssetCommissioned); err != nil {
		return err
	}

	// trigger node upgrade event
	go e.mgr.runActiveJob()

	return nil
}

// eventValidate perfoms the validations
func (e *updateEvent) eventValidate() error {
	var err error
	e._enodes, err = e.mgr.commonEventValidate(e.nodeNames)
	if err != nil {
		return err
	}

	if e.hostGroup != "" && !IsValidHostGroup(e.hostGroup) {
		return errored.Errorf("invalid host-group specified: %q", e.hostGroup)
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
			return errored.Errorf("Updating these nodes as worker will result in no master node in the cluster, make sure atleast one node is commissioned as master.")
		}
	}
	return nil
}

// pepareInventory prepares the inventory for update event.
func (e *updateEvent) pepareInventory() error {
	hosts := []*configuration.AnsibleHost{}
	for _, node := range e._enodes {
		host := node.Cfg.(*configuration.AnsibleHost)
		if e.hostGroup != "" {
			host.SetGroup(e.hostGroup)
		}
		hosts = append(hosts, host)
	}
	e._hosts = hosts

	return nil
}

// updateRunner is the job runner that runs a cleanup playbook followed by provision playbook
// on one or more nodes. In case of provision failure the cleanup playbook it run again.
func (e *updateEvent) updateRunner(cancelCh CancelChannel, jobLogs io.Writer) error {
	outReader, cancelFunc, errCh := e.mgr.configuration.Cleanup(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs); err != nil {
		log.Errorf("first cleanup failed. Error: %s", err)
		// XXX: is there a case where we should continue on error here?
		return err
	}
	outReader, cancelFunc, errCh = e.mgr.configuration.Configure(e._hosts, e.extraVars)
	cfgErr := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs)
	if cfgErr == nil {
		return nil
	}
	log.Errorf("configuration failed, starting cleanup. Error: %s", cfgErr)
	outReader, cancelFunc, errCh = e.mgr.configuration.Cleanup(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs); err != nil {
		log.Errorf("second cleanup failed. Error: %s", err)
	}

	//return the error status from provisioning
	return cfgErr
}
