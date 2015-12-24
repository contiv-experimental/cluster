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
	tbn       vagrantssh.TestbedNode
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
	validNodeName    = "cluster-node1-0"
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
	hosts := []vagrantssh.HostInfo{
		{
			Name:        "self",
			SSHAddr:     "127.0.0.1",
			SSHPort:     "22",
			User:        "vagrant",
			PrivKeyFile: fmt.Sprintf("%s/../demo/files/insecure_private_key", pwd),
		},
	}
	s.tb = &vagrantssh.Baremetal{}
	s.Assert(c, s.tb.Setup(hosts), IsNil)
	s.tbn = s.tb.GetNodes()[0]
	s.Assert(c, s.tbn, NotNil)
	// When a new vagrant setup comes up cluster-manager can take a bit to
	// come up as it waits on collins container to come up, which depending on
	// image download speed can take a while, so we wait for cluster-manager
	// to start with a long timeout here. This way we have this long wait only once.
	// XXX: we can alternatively save the collins container in the image and cut
	// this wait altogether.
	out, err := tutils.ServiceStartAndWaitForUp(s.tbn, "clusterm", 1200)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	//provide test ansible playbooks and restart cluster-mgr
	src := fmt.Sprintf("%s/../demo/files/cli_test/*", pwd)
	dst := "/etc/default/"
	out, err = s.tbn.RunCommandWithOutput(fmt.Sprintf("sudo cp -rf %s %s", src, dst))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = tutils.ServiceRestartAndWaitForUp(s.tbn, "clusterm", 30)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TearDownSuite(c *C) {
	// don't cleanup if stop-on-error is set
	if os.Getenv("CONTIV_SOE") != "" && s.failed {
		return
	}
	s.tbn = nil
	s.tb.Teardown()
}

func (s *CliTestSuite) SetUpTest(c *C) {
	if issue, ok := s.skipTests[c.TestName()]; ok {
		c.Skip(fmt.Sprintf("skipped due to known issue: %q", issue))
	}

	//cleanup an existing dummy file, if any that our test ansible will create. Ignore error, if any.
	file := dummyAnsibleFile
	out, err := s.tbn.RunCommandWithOutput(fmt.Sprintf("rm %s", file))
	c.Logf("dummy file cleanup. Error: %s, Output: %s", err, out)
	// XXX: we cleanup up assets from collins instead of restarting it to save test time.
	// Ignore errors here as asset might not exist.
	out, err = s.tbn.RunCommandWithOutput(fmt.Sprintf(`curl --basic -u blake:admin:first -d status="Decommissioned" -d reason="test" -X POST http://localhost:9000/api/asset/%s`, validNodeName))
	c.Logf("asset decommission result: %s. Output: %s", err, out)
	out, err = s.tbn.RunCommandWithOutput(fmt.Sprintf(`curl --basic -u blake:admin:first -d reason=test -X DELETE http://localhost:9000/api/asset/%s`, validNodeName))
	c.Logf("asset deletion result: %s. Output: %s", err, out)
	out, err = tutils.ServiceRestartAndWaitForUp(s.tbn, "clusterm", 90)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	c.Logf("clusterm is running. %s", out)
}

func (s *CliTestSuite) TearDownTest(c *C) {
	if s.failed {
		out, _ := tutils.ServiceLogs(s.tbn, "clusterm", 100)
		c.Logf(out)
	}

	// don't cleanup and stop the tests immediately if stop-on-error is set
	if os.Getenv("CONTIV_SOE") != "" && s.failed {
		c.Fatalf("%s failed. Stopping the tests as stop on error was set. Please check test logs to determine the actual failure. The system is left in same state for debugging.", c.TestName())
	}

	out, err := tutils.ServiceStop(s.tbn, "clusterm")
	c.Check(err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TestCommissionNonExistentNode(c *C) {
	nodeName := invalidNodeName
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
	exptStr := fmt.Sprintf(".*asset.*%s.*doesn't exists.*", nodeName)
	// XXX: somehow the following checker doesn't match the expression,
	// so resorting to a regex check here.
	//s.Assert(c, out, Matches, exptStr, Commentf("output: %s", out))
	if match, err := regexp.MatchString(exptStr, out); err != nil || !match {
		s.Assert(c, false, Equals, true, Commentf("output: %s", out))
	}
}

func (s *CliTestSuite) TestCommissionDisappearedNode(c *C) {
	nodeName := validNodeName
	// stop serf discovery
	out, err := tutils.ServiceStop(s.tbn, "serf")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	defer func() {
		// start serf discovery
		out, err := tutils.ServiceStart(s.tbn, "serf")
		s.Assert(c, err, IsNil, Commentf("output: %s", out))
	}()
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err = s.tbn.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, ErrorMatches, "node has disappeared", Commentf("output: %s", out))
}

func checkProvisionStatus(tbn vagrantssh.TestbedNode, nodeName, exptdStatus string) (string, error) {
	exptdStr := fmt.Sprintf(`.*"status".*"%s".*`, exptdStatus)
	return tutils.WaitForDone(func() (string, bool) {
		cmdStr := fmt.Sprintf("clusterctl node get %s", nodeName)
		out, err := tbn.RunCommandWithOutput(cmdStr)
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
	out, err := s.tbn.RunCommandWithOutput(fmt.Sprintf("sudo mv %s %s", src, dst))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	defer func() {
		out, err := s.tbn.RunCommandWithOutput(fmt.Sprintf("sudo mv %s %s", dst, src))
		s.Assert(c, err, IsNil, Commentf("output: %s", out))
	}()

	nodeName := validNodeName
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err = s.tbn.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = checkProvisionStatus(s.tbn, nodeName, "Unallocated")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TestCommissionSuccess(c *C) {
	nodeName := validNodeName

	// provision the node
	cmdStr := fmt.Sprintf("clusterctl node commission %s", nodeName)
	out, err := s.tbn.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = checkProvisionStatus(s.tbn, nodeName, "Allocated")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	// verify that site.yml got executed on the node and created the dummy file
	file := dummyAnsibleFile
	out, err = s.tbn.RunCommandWithOutput(fmt.Sprintf("stat -t %s", file))
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
}

func (s *CliTestSuite) TestDecommissionSuccess(c *C) {
	nodeName := validNodeName

	//commision the node
	s.TestCommissionSuccess(c)

	// decommission the node
	cmdStr := fmt.Sprintf("clusterctl node decommission %s", nodeName)
	out, err := s.tbn.RunCommandWithOutput(cmdStr)
	s.Assert(c, err, IsNil, Commentf("output: %s", out))
	out, err = checkProvisionStatus(s.tbn, nodeName, "Decommissioned")
	s.Assert(c, err, IsNil, Commentf("output: %s", out))

	// verify that cleanup.yml got executed on the node and deleted the dummy file
	file := dummyAnsibleFile
	out, err = s.tbn.RunCommandWithOutput(fmt.Sprintf("stat -t %s", file))
	s.Assert(c, err, NotNil, Commentf("output: %s", out))
}
