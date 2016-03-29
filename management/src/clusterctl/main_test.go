// +build unittest

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

func (s *mainSuite) TestCommandArgValidationError(c *C) {
	tests := map[string]struct {
		f        func(*manager.Client, string, string) error
		args     []string
		exptdErr error
	}{
		"commission": {
			f:        nodeCommission,
			args:     []string{"", ""},
			exptdErr: errNodeNameMissing("commission"),
		},
		"decommission": {
			f:        nodeDecommission,
			args:     []string{"", ""},
			exptdErr: errNodeNameMissing("decommission"),
		},
		"maintenance": {
			f:        nodeMaintenance,
			args:     []string{"", ""},
			exptdErr: errNodeNameMissing("maintenance"),
		},
		"discover": {
			f:        nodeDiscover,
			args:     []string{"", ""},
			exptdErr: errNodeAddrMissing("discover"),
		},
		"discover_invalid_ip": {
			f:        nodeDiscover,
			args:     []string{"1.2.3.4.5", ""},
			exptdErr: errInvalidIPAddr("1.2.3.4.5"),
		},
	}

	for key, test := range tests {
		err := test.f(nil, test.args[0], test.args[1])
		c.Assert(err.Error(), Equals, test.exptdErr.Error(), Commentf("test key: %s", key))
	}
}
