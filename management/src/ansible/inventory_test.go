package ansible

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type ansibleSuite struct {
}

var _ = Suite(&ansibleSuite{})

func MatchFile(c *C, f *os.File, expContents string) {
	defer os.Remove(f.Name())
	fileContents, err := ioutil.ReadFile(f.Name())
	c.Assert(err, IsNil)
	c.Assert(string(fileContents), Equals, expContents)
}

func (s *ansibleSuite) TestInventoryFile(c *C) {
	group1 := "g1"
	group2 := "g2"
	host1 := "h1"
	host2 := "h2"
	addr1 := "a1"
	addr2 := "a2"
	var1 := "foo1"
	val1 := "bar1"
	var2 := "foo2"
	val2 := "bar2"

	hosts := []InventoryHost{
		NewInventoryHost(host1, addr1, group1, map[string]string{}),
		NewInventoryHost(host1, addr1, group1,
			map[string]string{
				var1: val1,
			}),
		NewInventoryHost(host2, addr2, group1,
			map[string]string{
				var1: val1,
				var2: val2,
			}),
		NewInventoryHost(host1, addr1, group2,
			map[string]string{
				var1: val1,
			}),
		NewInventoryHost(host2, addr2, group2,
			map[string]string{
				var1: val1,
				var2: val2,
			}),
	}

	// XXX: the inventory generated is pretty ugly but
	// hopefully we can address this once we get
	// https://github.com/golang/go/issues/9969 fixed in text/template.
	// for now we tweak the expected text string below to match what is generated.
	singleHostFile := fmt.Sprintf(`
[%s]
%s ansible_ssh_host=%s 


	`, group1, host1, addr1)
	i := NewInventory(hosts[:1])
	f, err := NewInventoryFile(i)
	c.Assert(err, IsNil)
	c.Assert(f, NotNil)
	MatchFile(c, f, singleHostFile)

	singleHostWithVarsFile := fmt.Sprintf(`
[%s]
%s ansible_ssh_host=%s  %s=%s 


	`, group1, host1, addr1, var1, val1)
	i = NewInventory(hosts[1:2])
	f, err = NewInventoryFile(i)
	c.Assert(err, IsNil)
	c.Assert(f, NotNil)
	MatchFile(c, f, singleHostWithVarsFile)

	multiHostWithVarsFile := fmt.Sprintf(`
[%s]
%s ansible_ssh_host=%s  %s=%s 
%s ansible_ssh_host=%s  %s=%s  %s=%s 


	`, group1, host1, addr1, var1, val1, host2, addr2, var1, val1, var2, val2)
	i = NewInventory(hosts[1:3])
	f, err = NewInventoryFile(i)
	c.Assert(err, IsNil)
	c.Assert(f, NotNil)
	MatchFile(c, f, multiHostWithVarsFile)

	multiHostWithVarsMultiGroupsFile := fmt.Sprintf(`
[%s]
%s ansible_ssh_host=%s  %s=%s 
%s ansible_ssh_host=%s  %s=%s  %s=%s 

[%s]
%s ansible_ssh_host=%s  %s=%s 
%s ansible_ssh_host=%s  %s=%s  %s=%s 


	`, group1, host1, addr1, var1, val1, host2, addr2, var1, val1, var2, val2,
		group2, host1, addr1, var1, val1, host2, addr2, var1, val1, var2, val2)
	i = NewInventory(hosts[1:5])
	f, err = NewInventoryFile(i)
	c.Assert(err, IsNil)
	c.Assert(f, NotNil)
	MatchFile(c, f, multiHostWithVarsMultiGroupsFile)
}
