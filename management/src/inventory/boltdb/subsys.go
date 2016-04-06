package boltdb

import (
	"github.com/contiv/cluster/management/src/boltdb"
	"github.com/contiv/cluster/management/src/inventory"
)

// NewBoltdbSubsys initializes and return an instance of boltdb based inventory subsystem
func NewBoltdbSubsys(config boltdb.Config) (*inventory.GeneralSubsys, error) {
	client, err := boltdb.NewClientFromConfig(config)
	if err != nil {
		return nil, err
	}
	subsys := inventory.NewGeneralSubsys(client)

	// restore any previously added hosts
	assets, err := client.GetAllAssets()
	if err != nil {
		return nil, err
	}
	assets1 := assets.([]boltdb.Asset)
	for _, asset := range assets1 {
		a := inventory.NewAssetWithState(client, asset.Name, inventory.AssetStatusVals[asset.Status],
			inventory.AssetStateVals[asset.State])
		subsys.RestoreAsset(asset.Name, a)
	}

	return subsys, nil
}
