package collins

import (
	"github.com/contiv/cluster/management/src/collins"
	"github.com/contiv/cluster/management/src/inventory"
	"github.com/contiv/errored"
)

// NewCollinsSubsys initializes and return an instance of collins based inventory Subsys
func NewCollinsSubsys(config collins.Config) (*inventory.GeneralSubsys, error) {
	cc := collins.NewClientFromConfig(config)
	sc := inventory.SubsysClient(cc)
	ci := inventory.NewGeneralSubsys(cc)

	// create the customs states in collins
	for state, desc := range inventory.StateDescription {
		if err := cc.CreateState(state.String(), desc, inventory.Any.String()); err != nil {
			return nil, errored.Errorf("failed to create state %q in collins. Error: %s", state, err)
		}
	}

	// restore any previously added hosts
	cas, err := cc.GetAllAssets()
	if err != nil {
		return nil, err
	}
	cas1 := cas.([]collins.Asset)
	for _, ca := range cas1 {
		a := inventory.NewAssetWithState(sc, ca.Tag, inventory.AssetStatusVals[ca.Status],
			inventory.AssetStateVals[ca.State.Name])
		ci.RestoreAsset(ca.Tag, a)
	}

	return ci, nil
}
