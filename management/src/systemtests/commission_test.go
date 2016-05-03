// +build systemtest

package systemtests

import (
	"fmt"
	"os"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestCommissionNodeFailureNonExistent(c *C) {
	nodeName := invalidNodeName
	cmdStr := fmt.Sprintf("clusterctl node commission %s --host-group %s", nodeName, ansibleMasterGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*node.*%s.*doesn't exists.*", nodeName)
	s.assertMatch(c, exptStr, out)
}

func (s *SystemTestSuite) TestCommissionNodesFailureDisappeared(c *C) {
	nodeNames := validNodeNames
	nodeName := validNodeNames[1]
	// make sure test node is visible in inventory
	s.getNodeInfoSuccess(c, nodeName)

	// stop serf discovery on test node
	s.stopSerf(c, s.tbn2)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn1, nodeName, "failed")

	//try to commission the nodes
	nodesStr := strings.Join(nodeNames, " ")
	cmdStr := fmt.Sprintf("clusterctl nodes commission %s --host-group %s", nodesStr, ansibleMasterGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*one or more nodes are not in discovered state.*%s.*", nodeName)
	s.assertMatch(c, exptStr, out)
}

func (s *SystemTestSuite) TestCommissionProvisionFailure(c *C) {
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
	// create the temporary file, which shall be deleted as part of cleanup on provision failure
	s.touchFileAndWaitForStatToSucceed(c, s.tbn1, dummyAnsibleFile)
	cmdStr := fmt.Sprintf("clusterctl node commission %s --host-group %s", nodeName, ansibleMasterGroupName)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	s.checkProvisionStatus(c, s.tbn1, nodeName, "Unallocated")
	//make sure the temporary file got deleted
	s.waitForStatToFail(c, s.tbn1, dummyAnsibleFile)
}

func (s *SystemTestSuite) TestCommissionNodesFailureAlreadyAllocated(c *C) {
	nodeName := validNodeNames[0]
	secondNode := validNodeNames[1]
	nodeNames := []string{secondNode, nodeName}

	// commission the first node
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)

	//try to commission the nodes
	nodesStr := strings.Join(nodeNames, " ")
	cmdStr := fmt.Sprintf("clusterctl nodes commission %s --host-group %s", nodesStr, ansibleMasterGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*failed to update %s.*transition from.*%s.*%s.*is not allowed.*", nodeName, "Allocated", "Provisioning")
	s.assertMatch(c, exptStr, out)
	// make sure the already commissioned node stays Allocated
	s.checkProvisionStatus(c, s.tbn1, nodeName, "Allocated")
	// make sure the additional node stays Unallocated
	s.checkProvisionStatus(c, s.tbn1, secondNode, "Unallocated")
}

func (s *SystemTestSuite) TestCommissionNodeSuccess(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)
}

func (s *SystemTestSuite) TestCommissionNodeSerialSuccess(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)
	s.checkHostGroup(c, nodeName, "service-master")

	// commission second node and make sure it is added as worker
	nodeName = validNodeNames[1]
	s.commissionNode(c, nodeName, ansibleWorkerGroupName, s.tbn2)
	s.checkHostGroup(c, nodeName, "service-worker")
}

func (s *SystemTestSuite) TestCommissionNodeSerialMastersSuccess(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)
	s.checkHostGroup(c, nodeName, "service-master")

	// commission second node
	nodeName = validNodeNames[1]
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)
	s.checkHostGroup(c, nodeName, "service-master")
}

func (s *SystemTestSuite) TestCommissionNodesWithoutHostGroupFailure(c *C) {
	nodeName := validNodeNames[0]

	cmdStr := fmt.Sprintf("clusterctl nodes commission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := ".*invalid or empty host-group specified.*"
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestCommissionNodeWithoutHostGroupFailure(c *C) {
	nodeName := validNodeNames[0]

	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := ".*invalid or empty host-group specified.*"
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestCommissionWorkerNodesWithoutMasterFailure(c *C) {
	nodeName := validNodeNames[0]

	// commission a worker node directly
	cmdStr := fmt.Sprintf("clusterctl node commission %s --host-group %s", nodeName, ansibleWorkerGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := ".*Cannot commission a worker node without existence of a master node in the cluster, make sure atleast one master node is commissioned.*"
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestCommissionWorkerNodesWithMasterDisappearedFailure(c *C) {
	nodeName := validNodeNames[0]
	secondNode := validNodeNames[1]

	s.commissionNode(c, secondNode, ansibleMasterGroupName, s.tbn2)
	s.checkHostGroup(c, secondNode, "service-master")

	// stop serf discovery on test node
	s.stopSerf(c, s.tbn2)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn1, secondNode, "failed")

	// commission second node
	cmdStr := fmt.Sprintf("clusterctl node commission %s --host-group %s", nodeName, ansibleWorkerGroupName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := ".*Cannot commission a worker node without existence of a master node in the cluster, make sure atleast one master node is commissioned.*"
	s.assertMatch(c, exptStr, out)
}

func (s *SystemTestSuite) TestCommissionNodesSuccess(c *C) {
	nodeNames := validNodeNames
	s.commissionNodes(c, nodeNames, ansibleMasterGroupName)
	// make sure all nodes got assigned as master
	for _, name := range nodeNames {
		s.checkHostGroup(c, name, "service-master")
	}
}
