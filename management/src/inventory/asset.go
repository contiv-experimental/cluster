package inventory

import (
	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/contiv/errored"
)

// StateDescription is description of various assets states
var StateDescription = map[AssetState]string{
	Unknown:     "Node is in unknown state. This is the first state before initialization.",
	Discovered:  "Node is alive and discovered in monitoring subsystem",
	Disappeared: "Node has disappeared from monitoring subsystem. Check for possible hardware or network issues",
}

var (
	errAssetExists    = func(tag string) error { return errored.Errorf("asset %q already exists", tag) }
	errAssetNotExists = func(tag string) error { return errored.Errorf("asset %q doesn't exists", tag) }
)

// AssetStatusVals maps the status strings to corresponding enumerated values
var AssetStatusVals = map[string]AssetStatus{
	Incomplete.String():     Incomplete,
	New.String():            New,
	Unallocated.String():    Unallocated,
	Provisioning.String():   Provisioning,
	Provisioned.String():    Provisioned,
	Allocated.String():      Allocated,
	Cancelled.String():      Cancelled,
	Decommissioned.String(): Decommissioned,
	Maintenance.String():    Maintenance,
}

// AssetStateVals maps the state strings to corresponding enumerated values
var AssetStateVals = map[string]AssetState{
	strings.ToUpper(Unknown.String()):     Unknown,
	strings.ToUpper(Discovered.String()):  Discovered,
	strings.ToUpper(Disappeared.String()): Disappeared,
}

var lifecycleStatus = map[AssetStatus]map[AssetStatus]bool{
	Incomplete: {
		Unallocated: true,
	},
	New: {},
	Unallocated: {
		Provisioning: true,
	},
	Provisioning: {
		Unallocated: true,
		Allocated:   true,
	},
	Provisioned: {},
	Allocated: {
		Cancelled:   true,
		Maintenance: true,
	},
	Cancelled: {
		Decommissioned: true,
	},
	Decommissioned: {
		Provisioning: true,
	},
	Maintenance: {
		Unallocated: true,
		Allocated:   true,
	},
}

var lifecycleStates = map[AssetStatus]map[AssetState]bool{
	Incomplete: {},
	New:        {},
	Unallocated: {
		Discovered:  true,
		Disappeared: true,
	},
	Provisioning: {
		Discovered:  true,
		Disappeared: true,
	},
	Provisioned: {},
	Allocated: {
		Discovered:  true,
		Disappeared: true,
	},
	Cancelled: {
		Discovered:  true,
		Disappeared: true,
	},
	Decommissioned: {
		Discovered:  true,
		Disappeared: true,
	},
	Maintenance: {
		Discovered:  true,
		Disappeared: true,
	},
}

// Asset denotes a host or vm that is managed by the inventory susystem
type Asset struct {
	client     SubsysClient
	name       string
	status     AssetStatus
	prevStatus AssetStatus
	state      AssetState
	prevState  AssetState
}

// NewAssetWithState creates a new asset in the inventory in a discovered state and returns it.
func NewAssetWithState(client SubsysClient, name string, status AssetStatus, state AssetState) *Asset {
	return &Asset{
		client:     client,
		name:       name,
		status:     status,
		prevStatus: Incomplete,
		state:      state,
		prevState:  Unknown,
	}
}

// NewAsset creates a new asset in the inventory in a discovered state and returns it.
func NewAsset(client SubsysClient, name string) (*Asset, error) {
	a := &Asset{
		client:     client,
		name:       name,
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}

	if err := a.client.CreateAsset(name, a.status.String()); err != nil {
		return nil, err
	}

	if err := a.client.SetAssetStatus(name, a.status.String(), a.state.String(), StateDescription[a.state]); err != nil {
		//XXX: should we delete the asset here?
		return nil, err
	}

	log.Debugf("created asset: %+v", a)
	return a, nil
}

// SetStatus updates the status and/or state of an asset in the inventory after
// performing lifecyslce related validations.
func (a *Asset) SetStatus(status AssetStatus, state AssetState) error {
	if a.status == status && a.state == state {
		log.Infof("asset already in status: %q and state: %q, no action required", status, state)
		return nil
	}

	if _, ok := lifecycleStatus[a.status][status]; !ok && a.status != status {
		return errored.Errorf("transition from %q to %q is not allowed", a.status, status)
	}

	if _, ok := lifecycleStates[status][state]; !ok {
		return errored.Errorf("%q is not a valid state when asset is in %q status", state, status)
	}

	if err := a.client.SetAssetStatus(a.name, status.String(), state.String(), StateDescription[state]); err != nil {
		return err
	}

	a.prevStatus = a.status
	a.prevState = a.state
	a.status = status
	a.state = state

	return nil
}

// GetStatus returns the current status and state of an asset.
func (a *Asset) GetStatus() (AssetStatus, AssetState) {
	return a.status, a.state
}

//GetTag returns the inventory tag of the asset
func (a *Asset) GetTag() string {
	return a.name
}

// MarshalJSON implements the json marshaller for asset. It is done this way
// than making the fields public inorder to safeguard against direct state interpolation.
func (a *Asset) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		PrevStatus string `json:"prev_status"`
		State      string `json:"state"`
		PrevState  string `json:"prev_state"`
	}{
		Name:       a.name,
		Status:     a.status.String(),
		PrevStatus: a.prevStatus.String(),
		State:      a.state.String(),
		PrevState:  a.prevState.String(),
	})
}
