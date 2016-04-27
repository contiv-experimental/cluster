// +build systemtest

package systemtests

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestDiscoverNodeFailureAlreadyExist(c *C) {
	nodeName := validNodeNames[0]
	nodeAddr := validNodeAddrs[0]
	cmdStr := fmt.Sprintf("clusterctl discover %s", nodeAddr)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := fmt.Sprintf(".*one or more nodes already exist.*%s.*%s.*", nodeName, nodeAddr)
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestDiscoverNodeSuccess(c *C) {
	nodeName := validNodeNames[1]
	nodeAddr := validNodeAddrs[1]

	// stop serf on test node
	s.stopSerf(c, s.tbn2)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn1, nodeName, "failed")

	//cleanup inventory and restart clusterm
	s.nukeNodesInInventory(c)
	s.restartClusterm(c, s.tbn1, 30)

	// make sure node is not visible in inventory
	s.getNodeInfoFailureNonExistentNode(c, nodeName)

	// run discover command
	cmdStr := fmt.Sprintf("clusterctl discover %s", nodeAddr)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn1, nodeName, "alive")

	// make sure node is now visible in inventory
	s.getNodeInfoSuccess(c, nodeName)
}
