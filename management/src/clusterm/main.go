package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
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

func readConfig(c *cli.Context) ([]byte, error) {
	defConfig := `{
		"serf" : {
			"Addr" : "127.0.0.1:7373"
		},
		"collins" : {
			"URL" : "http://localhost:9000",
			"User": "blake",
			"Password": "admin:first"
		},
		"ansible" : {
			"configure-playbook": "site.yml",
			"cleanup-playbook": "cleanup.yml",
			"upgrade-playbook": "rolling-upgrade.yml",
			"playbook-location": "/vagrant/vendor/configuration/ansible",
			"user": "vagrant",
			"priv_key_file": "/vagrant/management/src/demo/files/insecure_private_key"
		},
		"manager" : {
			"Addr" : "localhost:9876"
		}
	}`
	if !c.GlobalIsSet("config") {
		log.Debugf("no configuration was specified, starting with default.")
		return []byte(defConfig), nil
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
	return config, nil
}

func startDaemon(c *cli.Context) {
	// set log level
	level := c.GlobalGeneric("debug").(*logLevel)
	log.SetLevel(level.value)

	var (
		err    error
		config []byte
	)
	if config, err = readConfig(c); err != nil {
		log.Fatalf("failed to read configuration. Error: %s", err)
	}
	mgr, err := manager.NewManager(config)
	if err != nil {
		log.Fatalf("failed to initialize the manager. Error: %s", err)
	}

	// start manager's processing loop
	errCh := make(chan error)
	go mgr.Run(errCh)
	select {
	case err := <-errCh:
		log.Fatalf("encountered an error: %s", err)
	}
}
