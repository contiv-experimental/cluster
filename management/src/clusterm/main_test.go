package main

import (
	"testing"

	"github.com/contiv/cluster/management/src/clusterm/manager"
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
			"playbook-location" : "override-location"
		}
	}`)
	exptdDst := manager.DefaultConfig()
	exptdDst.Ansible.PlaybookLocation = "override-location"

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, IsNil)
	c.Assert(dst, DeepEquals, exptdDst)
}

func (s *mainSuite) TestMergeConfigInvalidJSON(c *C) {
	dst := manager.DefaultConfig()
	srcBytes := []byte(`{
		"ansible" : {
			"playbook-location" : "extra-comma-error",
		}
	}`)

	_, err := mergeConfig(dst, srcBytes)
	c.Assert(err, ErrorMatches, "failed to parse configuration. Error:.*")
}
