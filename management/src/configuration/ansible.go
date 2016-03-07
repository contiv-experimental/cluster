package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/contiv/cluster/management/src/ansible"
	"github.com/imdario/mergo"
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
	config          *AnsibleSubsysConfig
	globalExtraVars string
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

// GetTag return the ansible inventory name/tag for the host
func (h *AnsibleHost) GetTag() string {
	return h.tag
}

// GetGroup return the ansible inventory group/role for the host
func (h *AnsibleHost) GetGroup() string {
	return h.group
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
		config:          config,
		globalExtraVars: DefaultValidJSON,
	}
}

func mergeExtraVars(dst, src string) (string, error) {
	var (
		d map[string]interface{}
		s map[string]interface{}
	)

	if err := json.Unmarshal([]byte(dst), &d); err != nil {
		return "", fmt.Errorf("failed to unmarshal dest extra vars %q. Error: %v", dst, err)
	}
	if err := json.Unmarshal([]byte(src), &s); err != nil {
		return "", fmt.Errorf("failed to unmarshal src extra vars %q. Error: %v", src, err)
	}
	if err := mergo.MergeWithOverwrite(&d, &s); err != nil {
		return "", fmt.Errorf("failed to merge extra vars, dst: %q src: %q. Error: %v", dst, src, err)
	}
	o, err := json.Marshal(d)
	if err != nil {
		return "", fmt.Errorf("failed to marshal resulting extra vars %q. Error: %v", o, err)
	}

	return string(o), nil
}

func (a *AnsibleSubsys) ansibleRunner(nodes []*AnsibleHost, playbook, extraVars string) (io.Reader, chan error) {
	// make error channel buffered, so it doesn't block
	errCh := make(chan error, 1)

	iNodes := []ansible.InventoryHost{}
	for _, n := range nodes {
		iNodes = append(iNodes, ansible.NewInventoryHost(n.tag, n.addr, n.group, n.vars))
	}

	// Pick extra variables for ansible, if any.
	// Merge the variables with following precedence (top one taking higher precedence):
	// - variables specified per action (i.e. configure, cleanup, upgrade)
	// - variables specified as globals
	// - variables specified at configuration time
	vars := DefaultValidJSON
	vars, err := mergeExtraVars(vars, a.config.ExtraVariables)
	if err != nil {
		errCh <- err
		return nil, errCh
	}
	vars, err = mergeExtraVars(vars, a.globalExtraVars)
	if err != nil {
		errCh <- err
		return nil, errCh
	}
	vars, err = mergeExtraVars(vars, extraVars)
	if err != nil {
		errCh <- err
		return nil, errCh
	}

	runner := ansible.NewRunner(ansible.NewInventory(iNodes), playbook, a.config.User, a.config.PrivKeyFile, vars)
	r, w := io.Pipe()
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

// SetGlobals sets the extra vars at a ansible subsys level
func (a *AnsibleSubsys) SetGlobals(extraVars string) error {
	a.globalExtraVars = extraVars
	return nil
}

// GetGlobals return the value of extra vars at a ansible subsys level
func (a *AnsibleSubsys) GetGlobals() string {
	return a.globalExtraVars
}
