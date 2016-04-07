package inventory

// GeneralSubsys implements the inventory sub-system. It is instantiated using
// the New* methods of specific subsystems like collins, boltdb and so on
type GeneralSubsys struct {
	client SubsysClient
	assets map[string]*Asset
}

// NewGeneralSubsys returns a instance of GeneralSubsys initialized with a subsystem client
func NewGeneralSubsys(client SubsysClient) *GeneralSubsys {
	return &GeneralSubsys{
		client: client,
		assets: make(map[string]*Asset),
	}
}

// RestoreAsset makes the subsystem update asset info
func (ci *GeneralSubsys) RestoreAsset(name string, asset *Asset) error {
	if _, ok := ci.assets[name]; ok {
		return errAssetExists(name)
	}

	ci.assets[name] = asset

	return nil
}

//AddAsset adds an asset to collins in 'Discovered' status
func (ci *GeneralSubsys) AddAsset(name string) error {
	if _, ok := ci.assets[name]; ok {
		return errAssetExists(name)
	}

	host, err := NewAsset(ci.client, name)
	if err != nil {
		return err
	}
	ci.assets[name] = host

	return nil
}

//SetAssetDiscovered sets an asset state to discovered
func (ci *GeneralSubsys) SetAssetDiscovered(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	status, _ := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(status, Discovered)
}

//SetAssetDisappeared sets an asset state to disappeared
func (ci *GeneralSubsys) SetAssetDisappeared(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	status, _ := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(status, Disappeared)
}

//SetAssetProvisioning sets an asset state to provisioning
func (ci *GeneralSubsys) SetAssetProvisioning(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(Provisioning, state)
}

//SetAssetCommissioned sets an asset status to unallocated
func (ci *GeneralSubsys) SetAssetCommissioned(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.assets[name].GetStatus()
	// collins equivalent of commissioned status in allocated
	return ci.assets[name].SetStatus(Allocated, state)
}

//SetAssetCancelled sets an asset state to cancelled
func (ci *GeneralSubsys) SetAssetCancelled(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(Cancelled, state)
}

//SetAssetDecommissioned sets an asset status to decommissioned
func (ci *GeneralSubsys) SetAssetDecommissioned(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(Decommissioned, state)
}

//SetAssetInMaintenance sets an asset state to decommissioned
func (ci *GeneralSubsys) SetAssetInMaintenance(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(Maintenance, state)
}

//SetAssetUnallocated sets an asset status to unallocated
func (ci *GeneralSubsys) SetAssetUnallocated(name string) error {
	if _, ok := ci.assets[name]; !ok {
		return errAssetNotExists(name)
	}

	_, state := ci.assets[name].GetStatus()
	return ci.assets[name].SetStatus(Unallocated, state)
}

//GetAsset finds and returns the asset in inventory
func (ci *GeneralSubsys) GetAsset(name string) SubsysAsset {
	if a, ok := ci.assets[name]; ok {
		return a
	}
	return nil
}

//GetAllAssets returns all the assets in inventory
func (ci *GeneralSubsys) GetAllAssets() SubsysAssets {
	return ci.assets
}
