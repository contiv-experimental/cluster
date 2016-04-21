package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
)

// maintenanceEvent triggers the upgrade workflow
type maintenanceEvent struct {
	mgr       *Manager
	nodeNames []string
	extraVars string
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
	if e.mgr.activeJob != nil {
		return errActiveJob(e.mgr.activeJob.String())
	}

	if err := e.mgr.inventory.SetAssetInMaintenance(e.nodeNames[0]); err != nil {
		// XXX. Log this to inventory
		return err
	}

	// trigger node upgrade event
	e.mgr.activeJob = NewJob(
		e.upgradeRunner,
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

// upgradeRunner is the job runner that runs upgrade plabook on one or more nodes
func (e *maintenanceEvent) upgradeRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	node, err := e.mgr.findNode(e.nodeNames[0])
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeNames[0])
	}

	outReader, cancelFunc, errCh := e.mgr.configuration.Upgrade(
		configuration.SubsysHosts([]*configuration.AnsibleHost{
			e.mgr.nodes[e.nodeNames[0]].Cfg.(*configuration.AnsibleHost),
		}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		log.Errorf("upgrade failed. Error: %s", err)
		return err
	}
	return nil
}
