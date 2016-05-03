package manager

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
)

// maintenanceEvent triggers the upgrade workflow
type maintenanceEvent struct {
	mgr       *Manager
	nodeNames []string
	extraVars string

	_hosts  configuration.SubsysHosts
	_enodes map[string]*node
}

// newMaintenanceEvent creates and returns maintenanceEvent
func newMaintenanceEvent(mgr *Manager, nodeNames []string, extraVars string) *maintenanceEvent {
	return &maintenanceEvent{
		mgr:       mgr,
		nodeNames: nodeNames,
		extraVars: extraVars,
	}
}

func (e *maintenanceEvent) String() string {
	return fmt.Sprintf("maintenanceEvent: %v", e.nodeNames)
}

func (e *maintenanceEvent) process() error {
	// err shouldn't be redefined below
	var err error

	err = e.mgr.checkAndSetActiveJob(
		e.upgradeRunner,
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
	if e._enodes, err = e.mgr.commonEventValidate(e.nodeNames); err != nil {
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

// pepareInventory prepares the inventory
func (e *maintenanceEvent) pepareInventory() error {
	hosts := []*configuration.AnsibleHost{}
	for _, node := range e._enodes {
		hosts = append(hosts, node.Cfg.(*configuration.AnsibleHost))
	}
	e._hosts = hosts

	return nil
}

// upgradeRunner is the job runner that runs upgrade plabook on one or more nodes
func (e *maintenanceEvent) upgradeRunner(cancelCh CancelChannel, jobLogs io.Writer) error {
	outReader, cancelFunc, errCh := e.mgr.configuration.Upgrade(e._hosts, e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc, jobLogs); err != nil {
		log.Errorf("upgrade failed. Error: %s", err)
		return err
	}
	return nil
}
