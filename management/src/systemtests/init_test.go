// +build systemtest

package systemtests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tutils "github.com/contiv/systemtests-utils"
	"github.com/contiv/vagrantssh"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type SystemTestSuite struct {
	tb        vagrantssh.Testbed
	tbn1      vagrantssh.TestbedNode
	tbn2      vagrantssh.TestbedNode
	failed    bool
	skipTests map[string]string
}

var _ = Suite(&SystemTestSuite{
	// add tests to skip due to known issues here.
	// The key of the map is test name like SystemTestSuite.TestCommissionDisappearedNode
	// The value of the map is the github issue# or url tracking reason for skip
	skipTests: map[string]string{},
})

var (
	validNodeNames   = []string{"cluster-node1-0", "cluster-node2-0"}
	validNodeAddrs   = []string{}
	invalidNodeName  = "invalid-test-node"
	dummyAnsibleFile = "/tmp/yay"
	testDataDir      = os.Getenv("TESTDATA_DIR")
)

func (s *SystemTestSuite) SetUpSuite(c *C) {
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
	s.tbn1 = s.tb.GetNode("node1")
	s.Assert(c, s.tbn1, NotNil)
	s.tbn2 = s.tb.GetNode("node2")
	s.Assert(c, s.tbn2, NotNil)
	validNodeAddrs = nodeIPs
	// When a new vagrant setup comes up cluster-manager can take a bit to
	// come up as it waits on collins container to come up and start serving it's API.
	// This can take a while, so we wait for cluster-manager
	// to start with a long timeout here. This way we have this long wait only once.
	s.restartClusterm(c, s.tbn1, 1200)
	//provide test ansible playbooks and restart cluster-mgr
	s.Assert(c, testDataDir, Not(Equals), "", Commentf("test data directory can't be empty"))
	src := fmt.Sprintf("%s/../%s/*", pwd, testDataDir)
	dst := "/etc/default/clusterm/"
	out, err := s.tbn1.RunCommandWithOutput(fmt.Sprintf("sudo cp -rf %s %s", src, dst))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	s.restartClusterm(c, s.tbn1, 30)
}

func (s *SystemTestSuite) TearDownSuite(c *C) {
	// don't cleanup if stop-on-error is set
	if os.Getenv("CONTIV_SOE") != "" && s.failed {
		return
	}
	s.tbn1 = nil
	s.tbn2 = nil
	s.tb.Teardown()
}

func (s *SystemTestSuite) SetUpTest(c *C) {
	if issue, ok := s.skipTests[c.TestName()]; ok {
		c.Skip(fmt.Sprintf("skipped due to known issue: %q", issue))
	}

	//cleanup an existing dummy file, if any that our test ansible will create. Ignore error, if any.
	file := dummyAnsibleFile
	out, err := s.tbn1.RunCommandWithOutput(fmt.Sprintf("rm %s", file))
	c.Logf("dummy file cleanup. Error: %v, Output: %s", err, out)
	out, err = s.tbn2.RunCommandWithOutput(fmt.Sprintf("rm %s", file))
	c.Logf("dummy file cleanup. Error: %v, Output: %s", err, out)

	// make sure serf is running
	s.startSerf(c, s.tbn1)
	s.startSerf(c, s.tbn2)

	//cleanup inventory and restart clusterm
	s.nukeNodesInInventory(c)
	s.restartClusterm(c, s.tbn1, 30)
}

func (s *SystemTestSuite) TearDownTest(c *C) {
	if s.failed {
		out, _ := tutils.ServiceLogs(s.tbn1, "clusterm", 100)
		c.Logf(out)
	}

	// don't cleanup and stop the tests immediately if stop-on-error is set
	if os.Getenv("CONTIV_SOE") != "" && s.failed {
		c.Fatalf("%s failed. Stopping the tests as stop on error was set. Please check test logs to determine the actual failure. The system is left in same state for debugging.", c.TestName())
	}
}
