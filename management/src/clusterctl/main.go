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
			Value: "localhost:9876",
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
					Action:  doAction(newNodePostActioner(nodecommission)),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "extra-vars, e",
							Value: "",
							Usage: "extra vars for ansible configuration. This should be a quoted json string.",
						},
					},
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a node",
					Action:  doAction(newNodePostActioner(nodeDecommission)),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "extra-vars, e",
							Value: "",
							Usage: "extra vars for ansible configuration. This should be a quoted json string.",
						},
					},
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a node in maintenance",
					Action:  doAction(newNodePostActioner(nodeMaintenance)),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "extra-vars, e",
							Value: "",
							Usage: "extra vars for ansible configuration. This should be a quoted json string.",
						},
					},
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get node's status information",
					Action:  doAction(newNodeGetActioner(nodeGet)),
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
					Action:  doAction(newNodeGetActioner(nodesGet)),
				},
			},
		},
	}

	app.Run(os.Args)
}

type actioner interface {
	procFlags(*cli.Context)
	procArgs(*cli.Context)
	action(*manager.Client) error
}

func doAction(a actioner) func(*cli.Context) {
	return func(c *cli.Context) {
		cClient := manager.NewClient(c.GlobalString("url"))
		a.procArgs(c)
		a.procFlags(c)
		if err := a.action(cClient); err != nil {
			log.Fatalf(err.Error())
		}
	}
}

type nodePostActioner struct {
	nodeName  string
	extraVars string
	postCb    func(c *manager.Client, nodeName, extraVars string) error
}

func newNodePostActioner(postCb func(c *manager.Client, nodeName, extraVars string) error) *nodePostActioner {
	return &nodePostActioner{postCb: postCb}
}

func (npa *nodePostActioner) procFlags(c *cli.Context) {
	npa.extraVars = c.String("extra-vars")
}

func (npa *nodePostActioner) procArgs(c *cli.Context) {
	npa.nodeName = c.Args().First()
}

func (npa *nodePostActioner) action(c *manager.Client) error {
	return npa.postCb(c, npa.nodeName, npa.extraVars)
}

func nodecommission(c *manager.Client, nodeName, extraVars string) error {
	if nodeName == "" {
		return errNodeNameMissing("commission")
	}
	return c.PostNodeCommission(nodeName, extraVars)
}

func nodeDecommission(c *manager.Client, nodeName, extraVars string) error {
	if nodeName == "" {
		return errNodeNameMissing("decommission")
	}
	return c.PostNodeDecommission(nodeName, extraVars)
}

func nodeMaintenance(c *manager.Client, nodeName, extraVars string) error {
	if nodeName == "" {
		return errNodeNameMissing("decommission")
	}
	return c.PostNodeInMaintenance(nodeName, extraVars)
}

type nodeGetActioner struct {
	nodeName string
	getCb    func(c *manager.Client, nodeName string) error
}

func newNodeGetActioner(getCb func(c *manager.Client, nodeName string) error) *nodeGetActioner {
	return &nodeGetActioner{getCb: getCb}
}

func (nga *nodeGetActioner) procFlags(c *cli.Context) {
	return
}

func (nga *nodeGetActioner) procArgs(c *cli.Context) {
	nga.nodeName = c.Args().First()
}

func (nga *nodeGetActioner) action(c *manager.Client) error {
	return nga.getCb(c, nga.nodeName)
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
