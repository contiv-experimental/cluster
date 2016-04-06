package boltdb

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/contiv/errored"
)

const (
	assetsBucket = "assets"
)

// Config denotes the configuration for boltdb client
type Config struct {
	DBFile string `json:"dbfile"`
}

// DefaultConfig returns the default configuration values for the boltdb client
func DefaultConfig() Config {
	return Config{
		DBFile: "/etc/default/clusterm/clusterm.boltdb",
	}
}

// Asset denotes the asset related information as read and stroed in boltdb.
type Asset struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	State     string `json:"state"`
	StateDesc string `json:"state_desc"`
}

// Client denotes state for a boltdb client
type Client struct {
	db     *bolt.DB
	config Config
}

// NewClientFromConfig initializes and return boltdb client using specified configuration
func NewClientFromConfig(config Config) (*Client, error) {
	db, err := bolt.Open(config.DBFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(assetsBucket))
		return err
	}); err != nil {
		return nil, err
	}
	return &Client{
		config: config,
		db:     db,
	}, nil
}

// NewClient initializes and return boltdb client using default configuration
func NewClient() (*Client, error) {
	return NewClientFromConfig(DefaultConfig())
}

// CreateAsset creates an asset with specified tag, status and state
func (c *Client) CreateAsset(tag, status string) error {
	a := Asset{
		Name:   tag,
		Status: status,
	}

	val, err := json.Marshal(a)
	if err != nil {
		return errored.Errorf("failed to marshal. Error: %v", err)
	}

	if err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(assetsBucket))
		err := b.Put([]byte(tag), val)
		return err
	}); err != nil {
		return err
	}

	return nil
}

// GetAsset queries and returns an asset with specified tag
func (c *Client) GetAsset(tag string) (Asset, error) {
	var (
		val []byte
		a   Asset
	)

	if err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(assetsBucket))
		val = b.Get([]byte(tag))
		if val == nil {
			return errored.Errorf("No asset found for name: %s", tag)
		}
		return nil
	}); err != nil {
		return a, err
	}

	if err := json.Unmarshal(val, &a); err != nil {
		return a, err
	}

	return a, nil
}

// GetAllAssets queries and returns a all the assets
func (c *Client) GetAllAssets() (interface{}, error) {
	var (
		vals   [][]byte
		a      Asset
		assets []Asset
	)

	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(assetsBucket))
		b.ForEach(func(k, v []byte) error {
			vals = append(vals, v)
			return nil
		})
		return nil
	})

	for _, val := range vals {
		if err := json.Unmarshal(val, &a); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}

	return assets, nil
}

// CreateState is a noop for boltdb
func (c *Client) CreateState(name, description, status string) error {
	return nil
}

// AddAssetLog creates a log entry for an asset
func (c *Client) AddAssetLog(tag, mtype, message string) error {
	return errored.Errorf("not implemented")
}

// SetAssetStatus sets the status of an asset
func (c *Client) SetAssetStatus(tag, status, state, reason string) error {
	a, err := c.GetAsset(tag)
	if err != nil {
		return err
	}
	a.Status = status
	a.State = state
	a.StateDesc = reason

	val, err := json.Marshal(a)
	if err != nil {
		return errored.Errorf("failed to marshal. Error: %v", err)
	}

	if err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(assetsBucket))
		err := b.Put([]byte(tag), val)
		return err
	}); err != nil {
		return err
	}

	return nil
}
