// +build unittest

package manager

import (
	"strings"

	"github.com/contiv/cluster/management/src/boltdb"
	"github.com/contiv/cluster/management/src/collins"
	"github.com/contiv/cluster/management/src/configuration"
	. "gopkg.in/check.v1"
)

type configSuite struct {
}

var _ = Suite(&configSuite{})

func (s *configSuite) TestReadConfigSuccess(c *C) {
	config := &Config{}
	confStr := `{
		"ansible" : {
			"playbook_location" : "foo"
		}
	}`
	_, err := config.read(strings.NewReader(confStr))
	c.Assert(err, IsNil)
	c.Assert(config.Ansible.PlaybookLocation, Equals, "foo")
}

func (s *configSuite) TestReadConfigInvalidJSON(c *C) {
	config := &Config{}
	confStr := `{
		"ansible" : {
			"playbook_location" : "extra-comma-error",
		}
	}`
	_, err := config.read(strings.NewReader(confStr))
	c.Assert(err, NotNil)
}

func (s *configSuite) TestMergeConfigSuccess(c *C) {
	dst := DefaultConfig()
	src := &Config{
		Ansible: configuration.AnsibleSubsysConfig{PlaybookLocation: "override-location"},
	}
	exptdDst := DefaultConfig()
	exptdDst.Ansible.PlaybookLocation = "override-location"

	_, err := dst.MergeFromConfig(src)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
}

func (s *configSuite) TestMergeConfigSuccessNoInventory(c *C) {
	dst := DefaultConfig()
	src := &Config{}
	exptdDst := DefaultConfig()

	_, err := dst.MergeFromConfig(src)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.Collins, Equals, (*collins.Config)(nil))
	c.Assert(dst.Inventory.BoltDB, Equals, (*boltdb.Config)(nil))
}

func (s *configSuite) TestMergeConfigSuccessCollinsInventory(c *C) {
	dst := DefaultConfig()
	src := &Config{
		Inventory: inventorySubsysConfig{Collins: &collins.Config{}},
	}
	exptdDst := DefaultConfig()
	exptdDst.Inventory.Collins = &collins.Config{}

	_, err := dst.MergeFromConfig(src)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.BoltDB, Equals, (*boltdb.Config)(nil))
	c.Assert(dst.Inventory.Collins, DeepEquals, exptdDst.Inventory.Collins)
}

func (s *configSuite) TestMergeConfigSuccessBoltdbInventory(c *C) {
	dst := DefaultConfig()
	src := &Config{
		Inventory: inventorySubsysConfig{BoltDB: &boltdb.Config{}},
	}
	exptdDst := DefaultConfig()
	exptdDst.Inventory.BoltDB = &boltdb.Config{}

	_, err := dst.MergeFromConfig(src)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.BoltDB, DeepEquals, exptdDst.Inventory.BoltDB)
	c.Assert(dst.Inventory.Collins, Equals, (*collins.Config)(nil))
}
