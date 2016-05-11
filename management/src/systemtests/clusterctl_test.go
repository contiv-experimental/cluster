// +build systemtest

package systemtests

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestSetGetGlobalExtraVarsSuccess(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl global set -e '{\\\"foo\\\":\\\"bar\\\"}'`)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	cmdStr = fmt.Sprintf(`clusterctl global get`)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"foo":.*"bar".*`
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestSetGetGlobalExtraVarsFailureInvalidJSON(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl global set -e '{\\\"foo\\\":}'`)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := `.*Request URL: globals.*extra_vars.*should be a valid json.*`
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestGetNodeInfoFailureNonExistentNode(c *C) {
	s.getNodeInfoFailureNonExistentNode(c, invalidNodeName)
}

func (s *SystemTestSuite) TestGetNodeInfoSuccess(c *C) {
	s.getNodeInfoSuccess(c, validNodeNames[0])
}

func (s *SystemTestSuite) TestGetNodesInfoSuccess(c *C) {
	cmdStr := `clusterctl nodes get`
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"monitoring_state":.*`
	s.assertMultiMatch(c, exptdOut, out, 2)
	exptdOut = `.*"inventory_state":.*`
	s.assertMultiMatch(c, exptdOut, out, 2)
	exptdOut = `.*"configuration_state".*`
	s.assertMultiMatch(c, exptdOut, out, 2)
}
