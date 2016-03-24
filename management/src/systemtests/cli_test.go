// +build systemtest

package systemtests

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	tutils "github.com/contiv/systemtests-utils"
	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CliTestSuite struct {
	tb        vagrantssh.Testbed
	tbn1      vagrantssh.TestbedNode
	tbn2      vagrantssh.TestbedNode
	failed    bool
	skipTests map[string]string
}

var _ = Suite(&CliTestSuite{
	// add tests to skip due to know issues here. Please add the issue#
	// being used to track
	skipTests: map[string]string{
		"CliTestSuite.TestCommissionDisappearedNode": "https://github.com/contiv/cluster/issues/28",
	},
})

var (
	validNodeNames   = []string{"cluster-node1-0", "cluster-node2-0"}
	invalidNodeName  = "invalid-test-node"
	dummyAnsibleFile = "/tmp/yay"
)

// XXX: go-check doesn't pass the test's state to the test set/teardown functions.
// So we have no way to know if a test failed and take some approrpate action.
// This hack let's me do that for now.
func (s *CliTestSuite) Assert(c *C, obtained interface{}, checker Checker, args ...interface{}) {
	if c.Check(obtained, checker, args...) == false {
		s.failed = true
		c.FailNow()
	}
}

func (s *CliTestSuite) SetUpSuite(c *C) {
	pwd, err := os.Getwd()
	s.Assert(c, err, IsNil)

	// The testbed is passed comma separate list of node IPs
	envStr := os.Getenv("CONTIV_NODE_IPS")
	nodeIPs := strings.Split(envStr, ",")
	s.Assert(c, len(nodeIPs), Equals, 2,
		Commentf("testbed expects 2 nodes but %d were passed. Node IPs: %q",
			len(nodeIPs), os.Getenv("CONTIV_NODE_IPS")))

	hosts := []vagrantssh.HostInfo{
		{
			Name:        "node1",
			SSHAddr:     nodeIPs[0],
			SSHPort:     "22",
			User:        "vagrant",
			PrivKeyFile: fmt.Sprintf("%s/../demo/files/insecure_private_key", pwd),
		},
		{
			Name:        "node2",
			SSHAddr:     nodeIPs[1],
			SSHPort:     "22",
			User:        "vagrant",
			PrivKeyFile: fmt.Sprintf("%s/../demo/files/insecure_private_key", pwd),
		},
	}
	s.tb = &vagrantssh.Baremetal{}
	s.Assert(c, s.tb.Setup(hosts), IsNil)
	s.tbn1 = s.tb.GetNodes()[0]
	s.Assert(c, s.tbn1, NotNil)
	s.tbn2 = s.tb.GetNodes()[1]
	s.Assert(c, s.tbn2, NotNil)
	// When a new vagrant setup comes up cluster-manager can take a bit to
	// come up as it waits on collins container to come up, which depending on
	// image download speed can take a while, so we wait for cluster-manager
	// to start with a long timeout here. This way we have this long wait only once.
	// XXX: we can alternatively save the collins container in the image and cut
	// this wait altogether.
	out, err := tutils.ServiceStartAndWaitForUp(s.tbn1, "clusterm", 1200)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	//provide test ansible playbooks and restart cluster-mgr
	src := fmt.Sprintf("%s/../demo/files/cli_test/*", pwd)
	dst := "/etc/default/clusterm/"
	out, err = s.tbn1.RunCommandWithOutput(fmt.Sprintf("sudo cp -rf %s %s", src, dst))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = tutils.ServiceRestartAndWaitForUp(s.tbn1, "clusterm", 30)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TearDownSuite(c *C) {
	// don't cleanup if stop-on-error is set
	if os.Getenv("CONTIV_SOE") != "" && s.failed {
		return
	}
	s.tbn1 = nil
	s.tbn2 = nil
	s.tb.Teardown()
}

func (s *CliTestSuite) nukeNodeInCollins(c *C, nodeName string) {
	// Ignore errors here as asset might not exist.
	out, err := s.tbn1.RunCommandWithOutput(fmt.Sprintf(`curl --basic -u blake:admin:first -d status="Decommissioned" -d reason="test" -X POST http://localhost:9000/api/asset/%s`, nodeName))
	c.Logf("asset decommission result: %v. Output: %s", err, out)
	out, err = s.tbn1.RunCommandWithOutput(fmt.Sprintf(`curl --basic -u blake:admin:first -d reason=test -X DELETE http://localhost:9000/api/asset/%s`, nodeName))
	c.Logf("asset deletion result: %v. Output: %s", err, out)
}

func (s *CliTestSuite) SetUpTest(c *C) {
	if issue, ok := s.skipTests[c.TestName()]; ok {
		c.Skip(fmt.Sprintf("skipped due to known issue: %q", issue))
	}

	//cleanup an existing dummy file, if any that our test ansible will create. Ignore error, if any.
	file := dummyAnsibleFile
	out, err := s.tbn1.RunCommandWithOutput(fmt.Sprintf("rm %s", file))
	c.Logf("dummy file cleanup. Error: %v, Output: %s", err, out)
	out, err = s.tbn2.RunCommandWithOutput(fmt.Sprintf("rm %s", file))
	c.Logf("dummy file cleanup. Error: %v, Output: %s", err, out)

	// XXX: we cleanup up assets from collins instead of restarting it to save test time.
	for _, name := range validNodeNames {
		s.nukeNodeInCollins(c, name)
	}

	out, err = tutils.ServiceRestartAndWaitForUp(s.tbn1, "clusterm", 90)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	c.Logf("clusterm is running. %s", out)
}

func (s *CliTestSuite) TearDownTest(c *C) {
	if s.failed {
		out, _ := tutils.ServiceLogs(s.tbn1, "clusterm", 100)
		c.Logf(out)
	}

	// don't cleanup and stop the tests immediately if stop-on-error is set
	if os.Getenv("CONTIV_SOE") != "" && s.failed {
		c.Fatalf("%s failed. Stopping the tests as stop on error was set. Please check test logs to determine the actual failure. The system is left in same state for debugging.", c.TestName())
	}

	out, err := tutils.ServiceStop(s.tbn1, "clusterm")
	c.Check(err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TestCommissionNonExistentNode(c *C) {
	nodeName := invalidNodeName
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*asset.*%s.*doesn't exists.*", nodeName)
	s.assertMatch(c, exptStr, out)
}

func (s *CliTestSuite) TestCommissionDisappearedNode(c *C) {
	nodeName := validNodeNames[0]
	// stop serf discovery
	out, err := tutils.ServiceStop(s.tbn1, "serf")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	defer func() {
		// start serf discovery
		out, err := tutils.ServiceStart(s.tbn1, "serf")
		s.Assert(c, err, IsNil, Commentf("output: %s", out))
	}()
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err = s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, ErrorMatches, "node has disappeared", Commentf("output: %s", out))
}

func checkProvisionStatus(tbn1 vagrantssh.TestbedNode, nodeName, exptdStatus string) (string, error) {
	exptdStr := fmt.Sprintf(`.*"status".*"%s".*`, exptdStatus)
	return tutils.WaitForDone(func() (string, bool) {
		cmdStr := fmt.Sprintf("clusterctl node get %s", nodeName)
		out, err := tbn1.RunCommandWithOutput(cmdStr)
		if err != nil {
			return out, false
			//replace newline with empty string for regex to match properly
		} else if match, err := regexp.MatchString(exptdStr,
			strings.Replace(out, "\n", "", -1)); err == nil && match {
			return out, true
		}
		return out, false
	}, 30, fmt.Sprintf("node is still not in %q status", exptdStatus))
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
	out, err = checkProvisionStatus(s.tbn1, nodeName, "Unallocated")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) commissionNode(c *C, nodeName string, nut vagrantssh.TestbedNode) {
	// provision the node
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = checkProvisionStatus(s.tbn1, nodeName, "Allocated")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	// verify that site.yml got executed on the node and created the dummy file
	file := dummyAnsibleFile
	out, err = nut.RunCommandWithOutput(fmt.Sprintf("stat -t %s", file))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TestCommissionSuccess(c *C) {
	nodeName := validNodeNames[0]
	s.commissionNode(c, nodeName, s.tbn1)
}

func (s *CliTestSuite) decommissionNode(c *C, nodeName string, nut vagrantssh.TestbedNode) {
	// decommission the node
	cmdStr := fmt.Sprintf("clusterctl node decommission %s", nodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = checkProvisionStatus(s.tbn1, nodeName, "Decommissioned")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	// verify that cleanup.yml got executed on the node and deleted the dummy file
	file := dummyAnsibleFile
	out, err = nut.RunCommandWithOutput(fmt.Sprintf("stat -t %s", file))
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
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
	cmdStr := fmt.Sprintf(`clusterctl node get %s`, invalidNodeName)
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptdOut := fmt.Sprintf(`.*node with name.*%s.*doesn't exists.*`, invalidNodeName)
	s.assertMatch(c, exptdOut, out)
}

func (s *CliTestSuite) TestGetNodeInfoSuccess(c *C) {
	cmdStr := fmt.Sprintf(`clusterctl node get %s`, validNodeNames[0])
	out, err := s.tbn1.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	exptdOut := `.*"monitoring-state":.*`
	s.assertMultiMatch(c, exptdOut, out, 1)
	exptdOut = `.*"inventory-state":.*`
	s.assertMultiMatch(c, exptdOut, out, 1)
	exptdOut = `.*"configuration-state".*`
	s.assertMultiMatch(c, exptdOut, out, 1)
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

func (s *CliTestSuite) assertMatch(c *C, exptd, rcvd string) {
	// XXX: the `Matches` checker doesn't match the expression in a multi-line
	// output so resorting to a regex check here.
	if match, err := regexp.MatchString(exptd, rcvd); err != nil || !match {
		s.Assert(c, false, Equals, true, Commentf("output: %s", rcvd))
	}
}

func (s *CliTestSuite) assertMultiMatch(c *C, exptd, rcvd string, eMatchCount int) {
	r := regexp.MustCompile(fmt.Sprintf("(?m)%s", exptd))
	s.Assert(c, len(r.FindAllString(rcvd, -1)), Equals, eMatchCount, Commentf("output: %s", rcvd))
}
