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

// read parses the configuration from the specified reader
// On success, it also return the updated receiver configuration
func (c *Config) read(r io.Reader) (*Config, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errored.Errorf("failed to read config. Error: %v", err)
	}

	if err := json.Unmarshal(bytes, c); err != nil {
		return nil, errInvalidJSON("config", err)
	}

	return c, nil
}

// MergeFromConfig merges the specified configuration into the receiver configuration
// On success, it also return the updated receiver configuration
func (c *Config) MergeFromConfig(src *Config) (*Config, error) {
	if err := mergo.MergeWithOverwrite(c, src); err != nil {
		return nil, errored.Errorf("failed to merge configuration. Error: %s", err)
	}
	return c, nil
}

// MergeFromReader merges the configuration from the specified reader
// On success, it also return the updated receiver configuration
func (c *Config) MergeFromReader(r io.Reader) (*Config, error) {
	config, err := (&Config{}).read(r)
	if err != nil {
		return nil, err
	}

	return c.MergeFromConfig(config)
}
