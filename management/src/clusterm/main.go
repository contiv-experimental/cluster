package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
	"github.com/imdario/mergo"
)

type logLevel struct {
	value log.Level
}

func (l *logLevel) Set(value string) error {
	var err error
	if l.value, err = log.ParseLevel(value); err != nil {
		return err
	}
	return nil
}

func (l *logLevel) String() string {
	return l.value.String()
}

func (l *logLevel) usage() string {
	return fmt.Sprintf("debug trace level: %s, %s, %s, %s, %s or %s", log.PanicLevel,
		log.FatalLevel, log.ErrorLevel, log.WarnLevel, log.InfoLevel, log.DebugLevel)
}

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "cluster manager daemon"
	app.Flags = []cli.Flag{
		cli.GenericFlag{
			Name:  "debug",
			Value: &logLevel{value: log.DebugLevel},
			Usage: (&logLevel{}).usage(),
		},
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "read cluster manager's configuration from file. Use '-' to read configuration from stdin",
		},
	}
	app.Action = startDaemon

	app.Run(os.Args)
}

func mergeConfig(dst *manager.Config, srcJSON []byte) (*manager.Config, error) {
	src := &manager.Config{}
	if err := json.Unmarshal(srcJSON, src); err != nil {
		return nil, fmt.Errorf("failed to parse configuration. Error: %s", err)
	}

	if err := mergo.MergeWithOverwrite(dst, src); err != nil {
		return nil, fmt.Errorf("failed to merge configuration. Error: %s", err)
	}

	return dst, nil
}

func readConfig(c *cli.Context) (*manager.Config, error) {
	mgrConfig := manager.DefaultConfig()
	if !c.GlobalIsSet("config") {
		log.Debugf("no configuration was specified, starting with default.")
		return mgrConfig, nil
	}

	var (
		reader io.Reader
		err    error
		config []byte
	)
	if c.GlobalString("config") == "-" {
		log.Debugf("reading configuration from stdin")
		reader = bufio.NewReader(os.Stdin)
	} else {
		var f *os.File
		if f, err = os.Open(c.GlobalString("config")); err != nil {
			return nil, err
		}
		log.Debugf("reading configuration from file: %q", c.GlobalString("config"))
		reader = bufio.NewReader(f)
	}

	if config, err = ioutil.ReadAll(reader); err != nil {
		return nil, err
	}

	return mergeConfig(mgrConfig, config)
}

func startDaemon(c *cli.Context) {
	// set log level
	level := c.GlobalGeneric("debug").(*logLevel)
	log.SetLevel(level.value)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	var (
		err    error
		config *manager.Config
	)
	if config, err = readConfig(c); err != nil {
		log.Fatalf("failed to read configuration. Error: %s", err)
	}
	mgr, err := manager.NewManager(config)
	if err != nil {
		log.Fatalf("failed to initialize the manager. Error: %s", err)
	}

	// start manager's processing loop
	errCh := make(chan error, 5)
	go mgr.Run(errCh)
	select {
	case err := <-errCh:
		log.Fatalf("encountered an error: %s", err)
	}
}
