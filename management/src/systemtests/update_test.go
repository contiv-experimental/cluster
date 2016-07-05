// +build systemtest

package systemtests

import (
	"fmt"
	"os"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestUpdateNodesSuccessSameHostGroup(c *C) {
	s.commissionNodes(c, validNodeNames, ansibleMasterGroupName)
	s.updateNodes(c, validNodeNames, ansibleMasterGroupName)
}

func (s *SystemTestSuite) TestUpdateNodeSuccessChangeHostGroup(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNodes(c, validNodeNames, ansibleMasterGroupName)
	s.updateNode(c, nodeName, ansibleWorkerGroupName, s.tbn1)
}

func (s *SystemTestSuite) TestUpdateNodeFailureNoMasterLeft(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)

	cmdStr := fmt.Sprintf("clusterctl node update %s --host-group %s", nodeName, ansibleWorkerGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := ".*Updating these nodes as worker will result in no master node in the cluster, make sure atleast one node is commissioned as master.*"
	s.assertMatch(c, exptStr, out)
}

func (s *SystemTestSuite) TestUpdateNodeFailureUnallocatedNode(c *C) {
	nodeName := validNodeNames[0]

	cmdStr := fmt.Sprintf("clusterctl node update %s --host-group %s", nodeName, ansibleMasterGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*failed to update %s.*transition from.*%s.*%s.*is not allowed.*", nodeName, "Unallocated", "Maintenance")
	s.assertMatch(c, exptStr, out)
}

func (s *SystemTestSuite) TestUpdateNodeProvisionFailure(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)

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

	cmdStr := fmt.Sprintf("clusterctl node update %s --host-group %s", nodeName, ansibleMasterGroupName)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	s.checkProvisionStatus(c, s.tbn1, nodeName, "Unallocated")
	//make sure the temporary file get's deleted due to cleanup
	s.waitForStatToFail(c, s.tbn1, dummyAnsibleFile)
}
