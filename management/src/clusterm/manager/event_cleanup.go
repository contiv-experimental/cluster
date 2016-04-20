package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
)

type nodeCleanup struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

func newNodeCleanup(mgr *Manager, nodeName, extraVars string) *nodeCleanup {
	return &nodeCleanup{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeCleanup) String() string {
	return fmt.Sprintf("nodeCleanup: %s", e.nodeName)
}

func (e *nodeCleanup) process() error {
	node, err := e.mgr.findNode(e.nodeName)
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeName)
	}

	outReader, _, errCh := e.mgr.configuration.Cleanup(
		configuration.SubsysHosts([]*configuration.AnsibleHost{
			e.mgr.nodes[e.nodeName].Cfg.(*configuration.AnsibleHost),
		}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh); err != nil {
		log.Errorf("cleanup failed. Error: %s", err)
	}
	// set asset state to decommissioned
	if err := e.mgr.inventory.SetAssetDecommissioned(e.nodeName); err != nil {
		// XXX. Log this to collins
		return err
	}
	return nil
}
