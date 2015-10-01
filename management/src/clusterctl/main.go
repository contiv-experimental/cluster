package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
)

var errNodeNameMissing = func(c string) error { return fmt.Errorf("command %q expects a node name", c) }

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "utility to interact with cluster manager"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: "localhost:9999",
			Usage: "cluster manager's REST service url",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "node",
			Aliases: []string{"n"},
			Usage:   "node related operation",
			Subcommands: []cli.Command{
				{
					Name:    "commission",
					Aliases: []string{"c"},
					Usage:   "commission a node",
					Action:  doAction(nodeCommision),
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a node",
					Action:  doAction(nodeDecommision),
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a node in maintenance",
					Action:  doAction(nodeMaintenance),
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get node's status information",
					Action:  doAction(nodeGet),
				},
			},
		},
		{
			Name:    "nodes",
			Aliases: []string{"a"},
			Usage:   "all nodes related operation",
			Subcommands: []cli.Command{
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get status information for all nodes",
					Action:  doAction(nodesGet),
				},
			},
		},
	}

	app.Run(os.Args)
}

func doAction(cb func(*manager.Client, string) error) func(*cli.Context) {
	return func(c *cli.Context) {
		cClient := manager.NewClient(c.GlobalString("url"))
		if err := cb(cClient, c.Args().First()); err != nil {
			log.Fatalf(err.Error())
		}
	}
}

func nodeCommision(c *manager.Client, nodeName string) error {
	if nodeName == "" {
		return errNodeNameMissing("commission")
	}
	return c.PostNodeCommission(nodeName)
}

func nodeDecommision(c *manager.Client, nodeName string) error {
	if nodeName == "" {
		return errNodeNameMissing("decommission")
	}
	return c.PostNodeDecommission(nodeName)
}

func nodeMaintenance(c *manager.Client, nodeName string) error {
	if nodeName == "" {
		return errNodeNameMissing("decommission")
	}
	return c.PostNodeInMaintenance(nodeName)
}

func nodeGet(c *manager.Client, nodeName string) error {
	var (
		out []byte
		err error
	)

	if nodeName == "" {
		return errNodeNameMissing("get")
	}

	if out, err = c.GetNode(nodeName); err != nil {
		return err
	}

	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "\t")
	outBuf.WriteTo(os.Stdout)
	return nil
}

func nodesGet(c *manager.Client, noop string) error {
	var (
		out []byte
		err error
	)

	if out, err = c.GetAllNodes(); err != nil {
		return err
	}

	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "\t")
	outBuf.WriteTo(os.Stdout)
	return nil
}
