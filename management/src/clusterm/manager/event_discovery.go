package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
)

// nodeDiscover triggers the node discovery workflow
type nodeDiscover struct {
	mgr       *Manager
	nodeAddr  string
	extraVars string
}

// newNodeDiscover creates and returns nodeDiscover event
func newNodeDiscover(mgr *Manager, nodeAddr, extraVars string) *nodeDiscover {
	return &nodeDiscover{
		mgr:       mgr,
		nodeAddr:  nodeAddr,
		extraVars: extraVars,
	}
}

func (e *nodeDiscover) String() string {
	return fmt.Sprintf("nodeDiscover: %s", e.nodeAddr)
}

func (e *nodeDiscover) process() error {
	node, err := e.mgr.findNodeByMgmtAddr(e.nodeAddr)
	if err == nil {
		return errored.Errorf("a node %q already exists with the management address %q",
			node.Inv.GetTag(), e.nodeAddr)
	}

	// create a temporary ansible host config to provision the host in discover host-group
	hostCfg := configuration.NewAnsibleHost("node1", e.nodeAddr,
		ansibleDiscoverGroupName, map[string]string{
			ansibleNodeNameHostVar: "node1",
			ansibleNodeAddrHostVar: e.nodeAddr,
		})

	outReader, _, errCh := e.mgr.configuration.Configure(
		configuration.SubsysHosts([]*configuration.AnsibleHost{hostCfg}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh); err != nil {
		log.Errorf("discover failed. Error: %s", err)
		return err
	}
	return nil
}
