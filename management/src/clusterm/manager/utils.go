package manager

import (
	"github.com/contiv/cluster/management/src/inventory"
	"github.com/contiv/errored"
)

func nodeNotExistsError(nameOrAddr string) error {
	return errored.Errorf("node with name or address %q doesn't exists", nameOrAddr)
}

func nodeConfigNotExistsError(name string) error {
	return errored.Errorf("the configuration info for node %q doesn't exist", name)
}

func nodeInventoryNotExistsError(name string) error {
	return errored.Errorf("the inventory info for node %q doesn't exist", name)
}

func (m *Manager) findNode(name string) (*node, error) {
	n, ok := m.nodes[name]
	if !ok {
		return nil, nodeNotExistsError(name)
	}
	return n, nil
}

func (m *Manager) findNodeByMgmtAddr(addr string) (*node, error) {
	for _, node := range m.nodes {
		if node.Mon.GetMgmtAddress() == addr {
			return node, nil
		}
	}
	return nil, nodeNotExistsError(addr)
}

func (m *Manager) isMasterNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.Cfg == nil {
		return false, nodeConfigNotExistsError(name)
	}
	return n.Cfg.GetGroup() == ansibleMasterGroupName, nil
}

func (m *Manager) isWorkerNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.Cfg == nil {
		return false, nodeConfigNotExistsError(name)
	}
	return n.Cfg.GetGroup() == ansibleWorkerGroupName, nil
}

func (m *Manager) isDiscoveredNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.Inv == nil {
		return false, nodeInventoryNotExistsError(name)
	}
	_, state := n.Inv.GetStatus()
	return state == inventory.Discovered, nil
}

func (m *Manager) isDiscoveredAndAllocatedNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.Inv == nil {
		return false, nodeInventoryNotExistsError(name)
	}
	status, state := n.Inv.GetStatus()
	return state == inventory.Discovered && status == inventory.Allocated, nil
}
