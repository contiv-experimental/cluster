package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
)

// nodeInMaintenance triggers the upgrade workflow
type nodeInMaintenance struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

// newNodeInMaintenance creates and returns nodeInMaintenance event
func newNodeInMaintenance(mgr *Manager, nodeName, extraVars string) *nodeInMaintenance {
	return &nodeInMaintenance{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeInMaintenance) String() string {
	return fmt.Sprintf("nodeInMaintenance: %s", e.nodeName)
}

func (e *nodeInMaintenance) process() error {
	if e.mgr.activeJob != nil {
		return errActiveJob(e.mgr.activeJob.String())
	}

	if err := e.mgr.inventory.SetAssetInMaintenance(e.nodeName); err != nil {
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
				if err := e.mgr.inventory.SetAssetUnallocated(e.nodeName); err != nil {
					// XXX. Log this to inventory
					log.Errorf("failed to update state in inventory, Error: %v", err)
				}
				return
			}
			// set asset state to commissioned
			if err := e.mgr.inventory.SetAssetCommissioned(e.nodeName); err != nil {
				// XXX. Log this to inventory
				log.Errorf("failed to update state in inventory, Error: %v", err)
			}
		})
	go e.mgr.activeJob.Run()
	return nil
}

// upgradeRunner is the job runner that runs upgrade plabook on one or more nodes
func (e *nodeInMaintenance) upgradeRunner(cancelCh CancelChannel) error {
	// reset active job status once done
	defer func() { e.mgr.activeJob = nil }()

	node, err := e.mgr.findNode(e.nodeName)
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeName)
	}

	outReader, cancelFunc, errCh := e.mgr.configuration.Upgrade(
		configuration.SubsysHosts([]*configuration.AnsibleHost{
			e.mgr.nodes[e.nodeName].Cfg.(*configuration.AnsibleHost),
		}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh, cancelCh, cancelFunc); err != nil {
		log.Errorf("upgrade failed. Error: %s", err)
		return err
	}
	return nil
}
