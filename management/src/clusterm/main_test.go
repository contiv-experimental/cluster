// +build unittest

package main

import (
	"testing"

	"github.com/contiv/cluster/management/src/boltdb"
	"github.com/contiv/cluster/management/src/clusterm/manager"
	"github.com/contiv/cluster/management/src/collins"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type mainSuite struct {
}

var _ = Suite(&mainSuite{})

func (s *mainSuite) TestMergeConfigSuccess(c *C) {
	dst := manager.DefaultConfig()
	srcBytes := []byte(`{
		"ansible" : {
			"playbook_location" : "override-location"
		}
	}`)
	exptdDst := manager.DefaultConfig()
	exptdDst.Ansible.PlaybookLocation = "override-location"

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
}

func (s *mainSuite) TestMergeConfigSuccessNoInventory(c *C) {
	dst := manager.DefaultConfig()
	srcBytes := []byte(`{
	}`)
	exptdDst := manager.DefaultConfig()

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.Collins, Equals, (*collins.Config)(nil))
	c.Assert(dst.Inventory.BoltDB, Equals, (*boltdb.Config)(nil))

	dst = manager.DefaultConfig()
	srcBytes = []byte(`{
		"inventory" : {}
	}`)
	exptdDst = manager.DefaultConfig()

	_, err = mergeConfig(dst, srcBytes)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.Collins, Equals, (*collins.Config)(nil))
	c.Assert(dst.Inventory.BoltDB, Equals, (*boltdb.Config)(nil))
}

func (s *mainSuite) TestMergeConfigSuccessCollinsInventory(c *C) {
	dst := manager.DefaultConfig()
	srcBytes := []byte(`{
		"inventory" : {
			"collins" : {}
		}
	}`)
	exptdDst := manager.DefaultConfig()
	exptdDst.Inventory.Collins = &collins.Config{}

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.BoltDB, Equals, (*boltdb.Config)(nil))
}

func (s *mainSuite) TestMergeConfigSuccessBoltdbInventory(c *C) {
	dst := manager.DefaultConfig()
	srcBytes := []byte(`{
		"inventory" : {
			"boltdb" : {}
		}
	}`)
	exptdDst := manager.DefaultConfig()
	exptdDst.Inventory.BoltDB = &boltdb.Config{}

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
	c.Assert(dst.Inventory.Collins, Equals, (*collins.Config)(nil))
}

func (s *mainSuite) TestMergeConfigInvalidJSON(c *C) {
	dst := manager.DefaultConfig()
	srcBytes := []byte(`{
		"ansible" : {
			"playbook_location" : "extra-comma-error",
		}
	}`)

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, ErrorMatches, "failed to parse configuration. Error:.*")
}
