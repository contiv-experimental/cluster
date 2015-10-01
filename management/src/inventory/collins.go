package inventory

import (
	"fmt"

	"github.com/contiv/cluster/management/src/collins"
)

// CollinsSubsys implements the inventory sub-system for the collins inventory management database
type CollinsSubsys struct {
	client CollinsClient
	hosts  map[string]*Asset
}

// NewCollinsSubsys initializes and return an instance of CollinsSubsys
func NewCollinsSubsys(config *collins.Config) (*CollinsSubsys, error) {
	ci := &CollinsSubsys{
		client: collins.NewClientFromConfig(config),
		hosts:  make(map[string]*Asset),
	}

	// create the customs states in collins
	for state, desc := range description {
		if err := ci.client.CreateState(state.String(), desc, Any.String()); err != nil {
			return nil, fmt.Errorf("failed to create state %q in collins. Error: %s", state, err)
		}
	}

	// restore any previously added hosts
	cas, err := ci.client.GetAllAssets()
	if err != nil {
		return nil, err
	}
	for _, ca := range cas {
		var (
			a   *Asset
			err error
		)
		if a, err = NewAssetFromCollins(ci.client, ca.Tag); err != nil {
			return nil, fmt.Errorf("failed to restore host %q from collins. Error: %s", ca.Tag, err)
		}
		ci.hosts[ca.Tag] = a
	}

	return ci, nil
}

//AddAsset adds an asset to collins in 'Discovered' status
func (ci *CollinsSubsys) AddAsset(name string) error {
	if _, ok := ci.hosts[name]; ok {
		return errAssetExists(name)
	}

	var (
		host *Asset
		err  error
	)
	if host, err = NewAsset(ci.client, name); err != nil {
		return err
	}
	ci.hosts[name] = host

	return nil
}

//SetAssetDiscovered sets an asset state to discovered
func (ci *CollinsSubsys) SetAssetDiscovered(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	status, _ := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(status, Discovered)
}

//SetAssetDisappeared sets an asset state to disappeared
func (ci *CollinsSubsys) SetAssetDisappeared(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	status, _ := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(status, Disappeared)
}

//SetAssetProvisioning sets an asset state to provisioning
func (ci *CollinsSubsys) SetAssetProvisioning(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(Provisioning, state)
}

//SetAssetCommissioned sets an asset status to unallocated
func (ci *CollinsSubsys) SetAssetCommissioned(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.hosts[name].GetStatus()
	// collins equivalent of commissioned status in allocated
	return ci.hosts[name].SetStatus(Allocated, state)
}

//SetAssetCancelled sets an asset state to cancelled
func (ci *CollinsSubsys) SetAssetCancelled(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(Cancelled, state)
}

//SetAssetDecommissioned sets an asset status to decommissioned
func (ci *CollinsSubsys) SetAssetDecommissioned(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(Decommissioned, state)
}

//SetAssetInMaintenance sets an asset state to decommissioned
func (ci *CollinsSubsys) SetAssetInMaintenance(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(Maintenance, state)
}

//SetAssetUnallocated sets an asset status to unallocated
func (ci *CollinsSubsys) SetAssetUnallocated(name string) error {
	if _, ok := ci.hosts[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.hosts[name].GetStatus()
	return ci.hosts[name].SetStatus(Unallocated, state)
}

//GetAsset finds and returns the asset in inventory
func (ci *CollinsSubsys) GetAsset(name string) SubsysAsset {
	if a, ok := ci.hosts[name]; ok {
		return a
	}
	return nil
}

//GetAllAssets returns all the assets in inventory
func (ci *CollinsSubsys) GetAllAssets() SubsysAssets {
	return ci.hosts
}
