package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
)

type nodeUpgrade struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

func newNodeUpgrade(mgr *Manager, nodeName, extraVars string) *nodeUpgrade {
	return &nodeUpgrade{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeUpgrade) String() string {
	return fmt.Sprintf("nodeUpgrade: %s", e.nodeName)
}

func (e *nodeUpgrade) process() error {
	node, err := e.mgr.findNode(e.nodeName)
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeName)
	}

	outReader, _, errCh := e.mgr.configuration.Upgrade(
		configuration.SubsysHosts([]*configuration.AnsibleHost{
			e.mgr.nodes[e.nodeName].Cfg.(*configuration.AnsibleHost),
		}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh); err != nil {
		log.Errorf("upgrade failed. Error: %s", err)
		// set asset state to provision-failed
		if err1 := e.mgr.inventory.SetAssetUnallocated(e.nodeName); err1 != nil {
			// XXX. Log this to collins
			return err1
		}
		return err
	}
	// set asset state to commissioned
	if err := e.mgr.inventory.SetAssetCommissioned(e.nodeName); err != nil {
		// XXX. Log this to collins
		return err
	}
	return nil
}
