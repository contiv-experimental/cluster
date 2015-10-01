package inventory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/contiv/cluster/management/src/collins"
	"github.com/golang/mock/gomock"
)

func TestNewAsset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eAsset := &Asset{
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	mClient := NewMockCollinsClient(ctrl)
	mClient.EXPECT().CreateAsset(eAsset.name, eAsset.status.String())
	mClient.EXPECT().SetAssetStatus(eAsset.name, eAsset.status.String(),
		eAsset.state.String(), description[eAsset.state])
	if rAsset, err := NewAsset(mClient, eAsset.name); err != nil {
		t.Fatalf("new asset failed. Error: %s", err)
	} else if rAsset.name != eAsset.name ||
		rAsset.status != eAsset.status ||
		rAsset.state != eAsset.state ||
		rAsset.prevStatus != eAsset.prevStatus ||
		rAsset.prevState != eAsset.prevState {
		t.Fatalf("mismatching asset info. expctd: %+v rcvd: %+v", eAsset, rAsset)
	}
}

func TestNewAssetCreateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eAsset := &Asset{
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	mClient := NewMockCollinsClient(ctrl)
	mClient.EXPECT().CreateAsset(eAsset.name,
		eAsset.status.String()).Return(fmt.Errorf("test error"))
	if _, err := NewAsset(mClient, eAsset.name); err == nil {
		t.Fatalf("new asset succeeded, expected to fail.")
	}
}

func TestNewAssetSetStatusFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eAsset := &Asset{
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	mClient := NewMockCollinsClient(ctrl)
	mClient.EXPECT().CreateAsset(eAsset.name, eAsset.status.String())
	mClient.EXPECT().SetAssetStatus(eAsset.name, eAsset.status.String(),
		eAsset.state.String(), description[eAsset.state]).Return(fmt.Errorf("test error"))
	if _, err := NewAsset(mClient, eAsset.name); err == nil {
		t.Fatalf("new asset succeeded, expected to fail.")
	}
}

func TestNewAssetFromCollins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eAsset := &Asset{
		name:       "foo",
		status:     Provisioned,
		prevStatus: Incomplete,
		state:      Disappeared,
		prevState:  Unknown,
	}
	mClient := NewMockCollinsClient(ctrl)
	mClient.EXPECT().GetAsset(eAsset.name).Return(collins.Asset{
		Tag:    eAsset.name,
		Status: eAsset.status.String(),
		State: struct {
			Name string `json:"NAME"`
		}{
			Name: strings.ToUpper(eAsset.state.String()),
		},
	}, nil)
	if rAsset, err := NewAssetFromCollins(mClient, eAsset.name); err != nil {
		t.Fatalf("new asset failed. Error: %s", err)
	} else if rAsset.name != eAsset.name ||
		rAsset.status != eAsset.status ||
		rAsset.state != eAsset.state ||
		rAsset.prevStatus != eAsset.prevStatus ||
		rAsset.prevState != eAsset.prevState {
		t.Fatalf("mismatching asset info. expctd: %+v rcvd: %+v", eAsset, rAsset)
	}
}

func TestNewAssetFromCollinsGetFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eAsset := &Asset{
		name:       "foo",
		status:     Provisioned,
		prevStatus: Incomplete,
		state:      Disappeared,
		prevState:  Unknown,
	}
	mClient := NewMockCollinsClient(ctrl)
	mClient.EXPECT().GetAsset(eAsset.name).Return(collins.Asset{}, fmt.Errorf("test failure"))
	if _, err := NewAssetFromCollins(mClient, eAsset.name); err == nil {
		t.Fatalf("new asset succeeded, expected to fail")
	}
}

func TestSetStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mClient := NewMockCollinsClient(ctrl)
	asset := &Asset{
		client:     mClient,
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	eAsset := &Asset{
		name:       "foo",
		status:     Provisioning,
		prevStatus: Unallocated,
		state:      Disappeared,
		prevState:  Discovered,
	}
	mClient.EXPECT().SetAssetStatus(asset.name, eAsset.status.String(),
		eAsset.state.String(), description[eAsset.state])
	if err := asset.SetStatus(eAsset.status, eAsset.state); err != nil {
		t.Fatalf("set asset status failed. Error: %s", err)
	} else if asset.name != eAsset.name ||
		asset.status != eAsset.status ||
		asset.state != eAsset.state ||
		asset.prevStatus != eAsset.prevStatus ||
		asset.prevState != eAsset.prevState {
		t.Fatalf("mismatching asset info. expctd: %+v rcvd: %+v", eAsset, asset)
	}
}

func TestSetStatusNoTransition(t *testing.T) {
	asset := &Asset{
		client:     nil,
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	eAsset := &Asset{
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	if err := asset.SetStatus(eAsset.status, eAsset.state); err != nil {
		t.Fatalf("set asset status failed. Error: %s", err)
	} else if asset.name != eAsset.name ||
		asset.status != eAsset.status ||
		asset.state != eAsset.state ||
		asset.prevStatus != eAsset.prevStatus ||
		asset.prevState != eAsset.prevState {
		t.Fatalf("mismatching asset info. expctd: %+v rcvd: %+v", eAsset, asset)
	}
}

func TestSetStatusInvalidStatusTransition(t *testing.T) {
	asset := &Asset{
		client:     nil,
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	eAsset := &Asset{
		name:       "foo",
		status:     Cancelled,
		prevStatus: Unallocated,
		state:      Disappeared,
		prevState:  Discovered,
	}
	errStr := "is not allowed"
	if err := asset.SetStatus(eAsset.status, eAsset.state); err == nil {
		t.Fatalf("set asset status succeeded, expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("unexpected error. expctd: %s, rcvd: %s", errStr, err)
	}
}

func TestSetStatusUnallowedState(t *testing.T) {
	asset := &Asset{
		client:     nil,
		name:       "foo",
		status:     Incomplete,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	eAsset := &Asset{
		name:       "foo",
		status:     Incomplete,
		prevStatus: Unallocated,
		state:      Disappeared,
		prevState:  Discovered,
	}
	errStr := "is not a valid state when asset is in"
	if err := asset.SetStatus(eAsset.status, eAsset.state); err == nil {
		t.Fatalf("set asset status succeeded, expected to fail")
	} else if !strings.Contains(err.Error(), errStr) {
		t.Fatalf("unexpected error. expctd: %s, rcvd: %s", errStr, err)
	}
}

func TestSetStatusSetFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mClient := NewMockCollinsClient(ctrl)
	asset := &Asset{
		client:     mClient,
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	eAsset := &Asset{
		name:       "foo",
		status:     Unallocated,
		prevStatus: Incomplete,
		state:      Discovered,
		prevState:  Unknown,
	}
	mClient.EXPECT().SetAssetStatus(asset.name, Provisioning.String(),
		Discovered.String(), description[asset.state]).Return(fmt.Errorf("test failure"))
	if err := asset.SetStatus(Provisioning, Discovered); err == nil {
		t.Fatalf("set asset status succeeded expected to fail")
	} else if asset.name != eAsset.name ||
		asset.status != eAsset.status ||
		asset.state != eAsset.state ||
		asset.prevStatus != eAsset.prevStatus ||
		asset.prevState != eAsset.prevState {
		t.Fatalf("mismatching asset info. expctd: %+v rcvd: %+v", eAsset, asset)
	}
}
