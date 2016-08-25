package manager

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/contiv/cluster/management/src/boltdb"
	"github.com/contiv/cluster/management/src/collins"
	"github.com/contiv/cluster/management/src/configuration"
	"github.com/contiv/errored"
	"github.com/imdario/mergo"
	"github.com/mapuri/serf/client"
)

type clustermConfig struct {
	Addr string `json:"addr"`
}

type inventorySubsysConfig struct {
	Collins *collins.Config `json:"collins,omitempty"`
	BoltDB  *boltdb.Config  `json:"boltdb,omitempty"`
}

// Config is the configuration to cluster manager daemon
type Config struct {
	Serf      client.Config                     `json:"serf"`
	Inventory inventorySubsysConfig             `json:"inventory"`
	Ansible   configuration.AnsibleSubsysConfig `json:"ansible"`
	Manager   clustermConfig                    `json:"manager"`
}

// DefaultConfig returns the default configuration values for the cluster manager
// and it's sub-systems
func DefaultConfig() *Config {
	return &Config{
		Serf: client.Config{
			Addr: "127.0.0.1:7373",
		},
		Inventory: inventorySubsysConfig{
			BoltDB:  nil,
			Collins: nil,
		},
		Ansible: configuration.AnsibleSubsysConfig{
			ConfigurePlaybook: "site.yml",
			CleanupPlaybook:   "cleanup.yml",
			UpgradePlaybook:   "rolling-upgrade.yml",
			PlaybookLocation:  "/vagrant/vendor/ansible",
			User:              "vagrant",
			PrivKeyFile:       "/vagrant/management/src/demo/files/insecure_private_key",
		},
		Manager: clustermConfig{
			Addr: "0.0.0.0:9007",
		},
	}
}

// Read parses the configuration from the specified reader
func (c *Config) Read(r io.Reader) error {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, c); err != nil {
		return err
	}

	return nil
}

// Merge merges the passed configuration into the receiving configuration
func (c *Config) Merge(src *Config) error {
	if err := mergo.MergeWithOverwrite(c, src); err != nil {
		return errored.Errorf("failed to merge configuration. Error: %s", err)
	}
	return nil
}
