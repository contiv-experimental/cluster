package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/cluster/management/src/monitor"
)

// discoveredEvent processes the discovered event from monitoring subsystem
type discoveredEvent struct {
	mgr   *Manager
	nodes []monitor.SubsysNode
}

// newDiscoveredEvent creates and returns discoveredEvent event
func newDiscoveredEvent(mgr *Manager, nodes []monitor.SubsysNode) *discoveredEvent {
	return &discoveredEvent{
		mgr:   mgr,
		nodes: nodes,
	}
}

func (e *discoveredEvent) String() string {
	return fmt.Sprintf("discoveredEvent: %+v", e.nodes[0])
}

func (e *discoveredEvent) process() error {
	//XXX: need to form the name that adheres to collins tag requirements
	name := e.nodes[0].GetLabel() + "-" + e.nodes[0].GetSerial()

	enode, err := e.mgr.findNode(name)
	if err != nil && err.Error() == nodeNotExistsError(name).Error() {
		e.mgr.nodes[name] = &node{
			// XXX: node's role/group shall come from manager's role assignment logic or
			// from user configuration
			Cfg: configuration.NewAnsibleHost(name, e.nodes[0].GetMgmtAddress(),
				ansibleMasterGroupName, map[string]string{
					ansibleNodeNameHostVar: name,
					ansibleNodeAddrHostVar: e.nodes[0].GetMgmtAddress(),
				}),
		}
		enode = e.mgr.nodes[name]
	} else if err != nil {
		return err
	}

	// update node's monitoring info to the one received in the event
	enode.Mon = e.nodes[0]
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
