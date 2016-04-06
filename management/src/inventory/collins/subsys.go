package collins

import (
	"github.com/contiv/cluster/management/src/collins"
	"github.com/contiv/cluster/management/src/inventory"
	"github.com/contiv/errored"
)

// NewCollinsSubsys initializes and return an instance of collins based inventory Subsys
func NewCollinsSubsys(config collins.Config) (*inventory.GeneralSubsys, error) {
	client := collins.NewClientFromConfig(config)
	subsys := inventory.NewGeneralSubsys(client)

	// create the customs states in collins
	for state, desc := range inventory.StateDescription {
		if err := client.CreateState(state.String(), desc, inventory.Any.String()); err != nil {
			return nil, errored.Errorf("failed to create state %q in collins. Error: %s", state, err)
		}
	}

	// restore any previously added hosts
	assets, err := client.GetAllAssets()
	if err != nil {
		return nil, err
	}
	assets1 := assets.([]collins.Asset)
	for _, asset := range assets1 {
		a := inventory.NewAssetWithState(client, asset.Tag, inventory.AssetStatusVals[asset.Status],
			inventory.AssetStateVals[asset.State.Name])
		subsys.RestoreAsset(asset.Tag, a)
	}

	return subsys, nil
}
