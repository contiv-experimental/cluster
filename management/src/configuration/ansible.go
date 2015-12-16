package configuration

import (
	"io"
	"strings"

	"github.com/contiv/cluster/management/src/ansible"
)

// AnsibleSubsysConfig describes the configuration for ansible based configuration management subsystem
type AnsibleSubsysConfig struct {
	ConfigurePlaybook string `json:"configure-playbook"`
	CleanupPlaybook   string `json:"cleanup-playbook"`
	UpgradePlaybook   string `json:"upgrade-playbook"`
	PlaybookLocation  string `json:"playbook-location"`
	ExtraVariables    string `json:"extra-variables"`
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
	addr  string
	group string
	tag   string
	vars  map[string]string
}

// NewAnsibleHost instantiates and returns AnsibleHost
func NewAnsibleHost(tag, addr, group string, vars map[string]string) *AnsibleHost {
	return &AnsibleHost{
		tag:   tag,
		addr:  addr,
		group: group,
		vars:  vars,
	}
}

// SetVar sets a host variable value
func (h *AnsibleHost) SetVar(key, val string) {
	h.vars[key] = val
}

// SetGroup sets the host's group
func (h *AnsibleHost) SetGroup(group string) {
	h.group = group
}

// NewAnsibleSubsys instantiates and returns AnsibleSubsys
func NewAnsibleSubsys(config *AnsibleSubsysConfig) *AnsibleSubsys {
	return &AnsibleSubsys{
		config: config,
	}
}

func (a *AnsibleSubsys) ansibleRunner(nodes []*AnsibleHost, playbook, extraVars string) (io.Reader, chan error) {
	iNodes := []ansible.InventoryHost{}
	for _, n := range nodes {
		iNodes = append(iNodes, ansible.NewInventoryHost(n.tag, n.addr, n.group, n.vars))
	}
	// Pick extra variables for ansible, if any.
	// Precedence (top one taking higher precedence):
	// - variables specified per action (i.e. configure, cleanup, upgrade)
	// - varaibles as passed in configuration
	// XXX: would merging the variables be better/desirable instead?
	vars := `{"env": {}}`
	if strings.TrimSpace(extraVars) != "" {
		vars = strings.TrimSpace(extraVars)
	} else if strings.TrimSpace(a.config.ExtraVariables) != "" {
		vars = strings.TrimSpace(a.config.ExtraVariables)
	}
	runner := ansible.NewRunner(ansible.NewInventory(iNodes), playbook, a.config.User, a.config.PrivKeyFile, vars)
	r, w := io.Pipe()
	// make error channel buffered, so it doesn't block
	errCh := make(chan error, 1)
	go func(outStream io.Writer, errCh chan error) {
		defer r.Close()
		if err := runner.Run(outStream, outStream); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
		return
	}(w, errCh)
	return r, errCh
}

// Configure triggers the ansible playbook for configuration on specified nodes
func (a *AnsibleSubsys) Configure(nodes SubsysHosts, extraVars string) (io.Reader, chan error) {
	return a.ansibleRunner(nodes.([]*AnsibleHost), strings.Join([]string{a.config.PlaybookLocation,
		a.config.ConfigurePlaybook}, "/"), extraVars)
}

// Cleanup triggers the ansible playbook for cleanup on specified nodes
func (a *AnsibleSubsys) Cleanup(nodes SubsysHosts, extraVars string) (io.Reader, chan error) {
	return a.ansibleRunner(nodes.([]*AnsibleHost), strings.Join([]string{a.config.PlaybookLocation,
		a.config.CleanupPlaybook}, "/"), extraVars)
}

// Upgrade triggers the ansible playbook for upgrade on specified nodes
func (a *AnsibleSubsys) Upgrade(nodes SubsysHosts, extraVars string) (io.Reader, chan error) {
	return a.ansibleRunner(nodes.([]*AnsibleHost), strings.Join([]string{a.config.PlaybookLocation,
		a.config.UpgradePlaybook}, "/"), extraVars)
}
