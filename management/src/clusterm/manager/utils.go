package manager

import (
	"fmt"

	"github.com/contiv/cluster/management/src/inventory"
)

func nodeNotExistsError(name string) error {
	return fmt.Errorf("node with name %q doesn't exists", name)
}

func nodeConfigNotExistsError(name string) error {
	return fmt.Errorf("the configuration info for node %q doesn't exist", name)
}

func nodeInventoryNotExistsError(name string) error {
	return fmt.Errorf("the inventory info for node %q doesn't exist", name)
}

func (m *Manager) findNode(name string) (*node, error) {
	n, ok := m.nodes[name]
	if !ok {
		return nil, nodeNotExistsError(name)
	}
	return n, nil
}

func (m *Manager) isMasterNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.cInfo == nil {
		return false, nodeConfigNotExistsError(name)
	}
	return n.cInfo.GetGroup() == ansibleMasterGroupName, nil
}

func (m *Manager) isWorkerNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.cInfo == nil {
		return false, nodeConfigNotExistsError(name)
	}
	return n.cInfo.GetGroup() == ansibleWorkerGroupName, nil
}

func (m *Manager) isDiscoveredAndAllocatedNode(name string) (bool, error) {
	n, err := m.findNode(name)
	if err != nil {
		return false, err
	}
	if n.iInfo == nil {
		return false, nodeInventoryNotExistsError(name)
	}
	status, state := n.iInfo.GetStatus()
	return state == inventory.Discovered && status == inventory.Allocated, nil
}
