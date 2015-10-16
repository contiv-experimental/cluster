package configuration

import (
	"bytes"
	"io"
	"strings"

	"github.com/contiv/cluster/management/src/ansible"
)

// AnsibleSubsysConfig describes the configuration for ansible based configuration management subsystem
type AnsibleSubsysConfig struct {
	ConfigurePlabook string `json:"configure-playbook"`
	CleanupPlaybook  string `json:"cleanup-playbook"`
	UpgradePlaybook  string `json:"upgrade-playbook"`
	PlaybookLocation string `json:"playbook-location"`
	// XXX: revisit the user credential configuration. We may need to allow other provisions.
	User        string `json:"user"`
	PrivKeyFile string `json:"priv_key_file"`
}

// AnsibleSubsys implements the configuration subsystem based on ansible
type AnsibleSubsys struct {
	config *AnsibleSubsysConfig
}

// AnsibleHost describes host related info relevant for ansible inventory
type AnsibleHost struct {
	addr string
	role string
	tag  string
	vars map[string]string
}

// NewAnsibleHost instantiates and returns AnsibleHost
func NewAnsibleHost(tag, addr, role string, vars map[string]string) *AnsibleHost {
	return &AnsibleHost{
		tag:  tag,
		addr: addr,
		role: role,
		vars: vars,
	}
}

// SetVar sets a host variable value
func (h *AnsibleHost) SetVar(key, val string) {
	h.vars[key] = val
}

// NewAnsibleSubsys instantiates and returns AnsibleSubsys
func NewAnsibleSubsys(config *AnsibleSubsysConfig) *AnsibleSubsys {
	return &AnsibleSubsys{
		config: config,
	}
}

func (a *AnsibleSubsys) ansibleRunner(nodes []*AnsibleHost, playbook string) (io.Reader, io.Reader, chan error) {
	iNodes := []ansible.InventoryHost{}
	for _, n := range nodes {
		iNodes = append(iNodes, ansible.NewInventoryHost(n.tag, n.addr, n.role, n.vars))
	}
	runner := ansible.NewRunner(ansible.NewInventory(iNodes), playbook, a.config.User, a.config.PrivKeyFile)
	stdout := []byte{}
	stderr := []byte{}
	errCh := make(chan error)
	go func(stdout []byte, stderr []byte, errCh chan error) {
		var err error
		if stdout, stderr, err = runner.Run(); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
		return
	}(stdout, stderr, errCh)
	return bytes.NewReader(stdout), bytes.NewReader(stderr), errCh
}

// Configure triggers the ansible playbook for configuration on specified nodes
func (a *AnsibleSubsys) Configure(nodes SubsysHosts) (io.Reader, io.Reader, chan error) {
	return a.ansibleRunner(nodes.([]*AnsibleHost), strings.Join([]string{a.config.PlaybookLocation,
		a.config.ConfigurePlabook}, "/"))
}

// Cleanup triggers the ansible playbook for cleanup on specified nodes
func (a *AnsibleSubsys) Cleanup(nodes SubsysHosts) (io.Reader, io.Reader, chan error) {
	return a.ansibleRunner(nodes.([]*AnsibleHost), strings.Join([]string{a.config.PlaybookLocation,
		a.config.CleanupPlaybook}, "/"))
}

// Upgrade triggers the ansible playbook for upgrade on specified nodes
func (a *AnsibleSubsys) Upgrade(nodes SubsysHosts) (io.Reader, io.Reader, chan error) {
	return a.ansibleRunner(nodes.([]*AnsibleHost), strings.Join([]string{a.config.PlaybookLocation,
		a.config.UpgradePlaybook}, "/"))
}
