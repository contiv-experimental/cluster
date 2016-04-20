package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/cluster/management/src/monitor"
)

// nodeDiscovered processes the discovered event from monitoring subsytem
type nodeDiscovered struct {
	mgr  *Manager
	node monitor.SubsysNode
}

// newNodeDiscovered creates and returns nodeDiscovered event
func newNodeDiscovered(mgr *Manager, node monitor.SubsysNode) *nodeDiscovered {
	return &nodeDiscovered{
		mgr:  mgr,
		node: node,
	}
}

func (e *nodeDiscovered) String() string {
	return fmt.Sprintf("nodeDiscovered: %+v", e.node)
}

func (e *nodeDiscovered) process() error {
	//XXX: need to form the name that adheres to collins tag requirements
	name := e.node.GetLabel() + "-" + e.node.GetSerial()

	enode, err := e.mgr.findNode(name)
	if err != nil && err.Error() == nodeNotExistsError(name).Error() {
		e.mgr.nodes[name] = &node{
			// XXX: node's role/group shall come from manager's role assignment logic or
			// from user configuration
			Cfg: configuration.NewAnsibleHost(name, e.node.GetMgmtAddress(),
				ansibleMasterGroupName, map[string]string{
					ansibleNodeNameHostVar: name,
					ansibleNodeAddrHostVar: e.node.GetMgmtAddress(),
				}),
		}
		enode = e.mgr.nodes[name]
	} else if err != nil {
		return err
	}

	// update node's monitoring info to the one received in the event
	enode.Mon = e.node
	enode.Inv = e.mgr.inventory.GetAsset(name)
	if enode.Inv == nil {
		if err := e.mgr.inventory.AddAsset(name); err != nil {
			// XXX. Log this to collins
			log.Errorf("adding asset %q to discovered in inventory failed. Error: %s", name, err)
			return err
		}
		enode.Inv = e.mgr.inventory.GetAsset(name)
	} else if err := e.mgr.inventory.SetAssetDiscovered(name); err != nil {
		// XXX. Log this to collins
		log.Errorf("setting asset %q to discovered in inventory failed. Error: %s", name, err)
		return err
	}
	return nil
}
