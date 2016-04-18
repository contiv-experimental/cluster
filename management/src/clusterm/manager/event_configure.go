package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
)

type nodeConfigure struct {
	mgr       *Manager
	nodeName  string
	extraVars string
}

func newNodeConfigure(mgr *Manager, nodeName, extraVars string) *nodeConfigure {
	return &nodeConfigure{
		mgr:       mgr,
		nodeName:  nodeName,
		extraVars: extraVars,
	}
}

func (e *nodeConfigure) String() string {
	return fmt.Sprintf("nodeConfigure: %s", e.nodeName)
}

func (e *nodeConfigure) process() error {
	node, err := e.mgr.findNode(e.nodeName)
	if err != nil {
		return err
	}

	if node.Cfg == nil {
		return nodeConfigNotExistsError(e.nodeName)
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
		if name == e.nodeName {
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
	outReader, _, errCh := e.mgr.configuration.Configure(
		configuration.SubsysHosts([]*configuration.AnsibleHost{hostInfo}), e.extraVars)
	if err := logOutputAndReturnStatus(outReader, errCh); err != nil {
		log.Errorf("configuration failed. Error: %s", err)
		// set asset state back to unallocated
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
