// +build systemtest

package systemtests

import (
	"fmt"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *SystemTestSuite) TestDecommissionNodeSuccess(c *C) {
	nodeName := validNodeNames[0]

	// commission the node
	s.commissionNode(c, nodeName, ansibleMasterGroupName, s.tbn1)

	// decommission the node
	s.decommissionNode(c, nodeName, s.tbn1)
}

func (s *SystemTestSuite) TestDecommissionNodeSuccessSerial(c *C) {
	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	// commission the nodes. First node is master, second node is worker
	s.commissionNode(c, nodeName1, ansibleMasterGroupName, s.tbn1)
	s.commissionNode(c, nodeName2, ansibleWorkerGroupName, s.tbn2)

	// decommission the node
	s.decommissionNode(c, nodeName2, s.tbn2)
	s.decommissionNode(c, nodeName1, s.tbn1)
}

func (s *SystemTestSuite) TestDecommissionNodesSuccess(c *C) {
	nodeNames := validNodeNames

	// commission the nodes
	s.commissionNodes(c, nodeNames, ansibleMasterGroupName)

	// decommission the nodes
	s.decommissionNodes(c, nodeNames)
}

func (s *SystemTestSuite) TestDecommissionNodeFailureRemainingWorker(c *C) {
	nodeName1 := validNodeNames[0]
	nodeName2 := validNodeNames[1]

	//commission the nodes. First node is master, second node is worker
	s.commissionNode(c, nodeName1, ansibleMasterGroupName, s.tbn1)
	s.commissionNode(c, nodeName2, ansibleWorkerGroupName, s.tbn2)

	// decommission the master node
	cmdStr := fmt.Sprintf("clusterctl node decommission %s", nodeName1)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := ".*decommissioning.*will leave only worker nodes.*all worker nodes are decommissioned before last master node.*"
	s.assertMatch(c, exptdOut, out)
}

func (s *SystemTestSuite) TestDecommissionNodesFailureDisappeared(c *C) {
	nodeNames := validNodeNames
	nodeName := validNodeNames[1]

	// commission the nodes
	s.commissionNodes(c, nodeNames, ansibleMasterGroupName)

	// make sure test node is visible in inventory
	s.getNodeInfoSuccess(c, nodeName)

	// stop serf discovery on test node
	s.stopSerf(c, s.tbn2)

	// wait for serf membership to update
	s.waitForSerfMembership(c, s.tbn1, nodeName, "failed")

	//try to decommission the nodes
	nodesStr := strings.Join(nodeNames, " ")
	cmdStr := fmt.Sprintf("clusterctl nodes decommission %s", nodesStr)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*one or more nodes are not in discovered state.*%s.*", nodeName)
	s.assertMatch(c, exptStr, out)
}

func (s *SystemTestSuite) TestDecommissionNodesFailureAlreadyDecommissioned(c *C) {
	nodeName := validNodeNames[0]
	secondNode := validNodeNames[1]
	nodeNames := []string{secondNode, nodeName}

	// commission the nodes
	s.commissionNodes(c, nodeNames, ansibleMasterGroupName)

	// decommission one node
	s.decommissionNode(c, nodeName, s.tbn1)

	//try to decommission all the nodes
	nodesStr := strings.Join(nodeNames, " ")
	cmdStr := fmt.Sprintf("clusterctl nodes decommission %s", nodesStr)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*failed to update %s.*transition from.*%s.*%s.*is not allowed.*", nodeName, "Decommissioned", "Cancelled")
	s.assertMatch(c, exptStr, out)
	// make sure the already decommissioned node stays Decommissioned
	s.checkProvisionStatus(c, s.tbn1, nodeName, "Decommissioned")
	// make sure the additional node stays Allocated
	s.checkProvisionStatus(c, s.tbn1, secondNode, "Allocated")
}
