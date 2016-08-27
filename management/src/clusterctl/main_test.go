// +build unittest

package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type mainSuite struct {
}

var _ = Suite(&mainSuite{})

func (s *mainSuite) TestCommandArgValidationError(c *C) {
	tests := map[string]struct {
		f        validateCallback
		args     []string
		exptdErr error
	}{
		"one-node-name": {
			f:        validateOneArg,
			args:     []string{"", ""},
			exptdErr: errUnexpectedArgCount("1", len([]string{"", ""})),
		},
		"multi-node-name": {
			f:        validateMultiNodeNames,
			args:     []string{},
			exptdErr: errUnexpectedArgCount(">=1", len([]string{})),
		},
		"multi-node-addrs": {
			f:        validateMultiNodeAddrs,
			args:     []string{},
			exptdErr: errUnexpectedArgCount(">=1", len([]string{})),
		},
		"invalid-addr": {
			f:        validateMultiNodeAddrs,
			args:     []string{"1.2.3.4.5", ""},
			exptdErr: errInvalidIPAddr("1.2.3.4.5"),
		},
	}

	for key, test := range tests {
		err := test.f(test.args)
		c.Assert(err.Error(), Equals, test.exptdErr.Error(), Commentf("test key: %s", key))
	}
}
