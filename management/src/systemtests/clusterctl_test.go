// +build systemtest

package systemtests

import (
	"fmt"
	"os"

	. "gopkg.in/check.v1"
)

func (s *CliTestSuite) TestCommissionNonExistentNode(c *C) {
	nodeName := invalidNodeName
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*node.*%s.*doesn't exists.*", nodeName)
	s.assertMatch(c, exptStr, out)
}

func (s *CliTestSuite) TestCommissionDisappearedNode(c *C) {
	nodeName := validNodeNames[1]
	// make sure test node is visible in inventory
	s.getNodeInfoSuccess(c, nodeName)

	// stop serf discovery on test node
	s.stopSerf(c, s.tbn2)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn1, nodeName, "failed")

	//try to commission the node
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*node.*%s.*has disappeared.*", nodeName)
	s.assertMatch(c, exptStr, out)
}

func (s *CliTestSuite) TestCommissionProvisionFailure(c *C) {
	// temporarily move the site.yml file to sitmulate a failure
	pwd, err := os.Getwd()
	s.Assert(c, err, IsNil)
	src := fmt.Sprintf("%s/../demo/files/site.yml", pwd)
	dst := fmt.Sprintf("%s/../demo/files/site.yml.1", pwd)
	out, err := s.tbn1.RunCommandWithOutput(fmt.Sprintf("sudo mv %s %s", src, dst))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	defer func() {
		out, err := s.tbn1.RunCommandWithOutput(fmt.Sprintf("sudo mv %s %s", dst, src))
		s.Assert(c, err, IsNil, Commentf("output: %s", out))
	}()

	nodeName := validNodeNames[0]
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	s.checkProvisionStatus(c, s.tbn1, nodeName, "Unallocated")
}

func (s *CliTestSuite) TestCommissionSuccess(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, s.tbn1)
}

func (s *CliTestSuite) TestDecommissionSuccess(c *C) {
	nodeName := validNodeNames[0]

	//commision the node
	s.commissionNode(c, nodeName, s.tbn1)

	// decommission the node
	s.decommissionNode(c, nodeName, s.tbn1)
}

func (s *CliTestSuite) TestDecommissionSuccessTwoNodes(c *C) {
	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	//commision the nodes. First node is master, second node is worker
	s.commissionNode(c, nodeName1, s.tbn1)
	s.commissionNode(c, nodeName2, s.tbn2)

	// decommission the node
	s.decommissionNode(c, nodeName2, s.tbn2)
	s.decommissionNode(c, nodeName1, s.tbn1)
}

func (s *CliTestSuite) TestDecommissionFailureRemainingWorkerNodes(c *C) {
	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	//commision the nodes. First node is master, second node is worker
	s.commissionNode(c, nodeName1, s.tbn1)
	s.commissionNode(c, nodeName2, s.tbn2)

	// decommission the master node
	cmdStr := fmt.Sprintf("clusterctl node decommission %s", nodeName1)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := fmt.Sprintf(".*%s.*is a master node and it can only be decommissioned after all worker nodes have been decommissioned.*", nodeName1)
	s.assertMatch(c, exptdOut, out)
}

func (s *CliTestSuite) TestDiscoverNodeAlreadyExistError(c *C) {
	nodeName := validNodeNames[0]
	nodeAddr := validNodeAddrs[0]
	cmdStr := fmt.Sprintf("clusterctl discover %s", nodeAddr)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := fmt.Sprintf("a node.*%s.*already exists with the management address.*%s.*", nodeName, nodeAddr)
	s.assertMatch(c, exptdOut, out)
}

func (s *CliTestSuite) TestDiscoverSuccess(c *C) {
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

func (s *CliTestSuite) TestSetGetGlobalExtraVarsSuccess(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl global set -e '{\\\"foo\\\":\\\"bar\\\"}'`)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	cmdStr = fmt.Sprintf(`clusterctl global get`)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"foo":.*"bar".*`
	s.assertMatch(c, exptdOut, out)
}

func (s *CliTestSuite) TestSetGetGlobalExtraVarsFailureInvalidJSON(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl global set -e '{\\\"foo\\\":}'`)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := `.*Request: globals.*extra_vars.*should be a valid json.*`
	s.assertMatch(c, exptdOut, out)
}

func (s *CliTestSuite) TestGetNodeInfoFailureNonExistentNode(c *C) {
	s.getNodeInfoFailureNonExistentNode(c, invalidNodeName)
}

func (s *CliTestSuite) TestGetNodeInfoSuccess(c *C) {
	s.getNodeInfoSuccess(c, validNodeNames[0])
}

func (s *CliTestSuite) TestGetNodesInfoSuccess(c *C) {
	cmdStr := `clusterctl nodes get`
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"monitoring-state":.*`
	s.assertMultiMatch(c, exptdOut, out, 2)
	exptdOut = `.*"inventory-state":.*`
	s.assertMultiMatch(c, exptdOut, out, 2)
	exptdOut = `.*"configuration-state".*`
	s.assertMultiMatch(c, exptdOut, out, 2)
}
