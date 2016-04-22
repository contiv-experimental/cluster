package main

import (
	"bytes"
	"encoding/json"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
	"github.com/contiv/errored"
)

var (
	errNodeNameMissing = func(c string) error { return errored.Errorf("command %q expects a node name", c) }
	errNodeAddrMissing = func(c string) error { return errored.Errorf("command %q expects a node IP address", c) }
	errInvalidIPAddr   = func(a string) error { return errored.Errorf("failed to parse ip address %q", a) }

	clustermFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: manager.DefaultConfig().Manager.Addr,
			Usage: "cluster manager's REST service url",
		},
	}

	extraVarsFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "extra-vars, e",
			Value: "",
			Usage: "extra vars for ansible configuration. This should be a quoted json string.",
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "utility to interact with cluster manager"
	app.Flags = clustermFlags
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
					Action:  doAction(newPostActioner(nodeCommission)),
					Flags:   extraVarsFlags,
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a node",
					Action:  doAction(newPostActioner(nodeDecommission)),
					Flags:   extraVarsFlags,
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a node in maintenance",
					Action:  doAction(newPostActioner(nodeMaintenance)),
					Flags:   extraVarsFlags,
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get node's status information",
					Action:  doAction(newGetActioner(nodeGet)),
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
					Action:  doAction(newGetActioner(nodesGet)),
				},
			},
		},
		{
			Name:    "global",
			Aliases: []string{"g"},
			Usage:   "set/get global info",
			Subcommands: []cli.Command{
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get global info",
					Action:  doAction(newGetActioner(globalsGet)),
				},
				{
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "set global info",
					Flags:   extraVarsFlags,
					Action:  doAction(newPostActioner(globalsSet)),
				},
			},
		},
		{
			Name:    "discover",
			Aliases: []string{"d"},
			Usage:   "provision a node for discovery",
			Action:  doAction(newPostActioner(nodeDiscover)),
			Flags:   extraVarsFlags,
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

type postActioner struct {
	nodeNameOrAddr string
	extraVars      string
	postCb         func(c *manager.Client, nodeNameOrAddr, extraVars string) error
}

func newPostActioner(postCb func(c *manager.Client, nodeNameOrAddr, extraVars string) error) *postActioner {
	return &postActioner{postCb: postCb}
}

func (npa *postActioner) procFlags(c *cli.Context) {
	npa.extraVars = c.String("extra-vars")
}

func (npa *postActioner) procArgs(c *cli.Context) {
	npa.nodeNameOrAddr = c.Args().First()
}

func (npa *postActioner) action(c *manager.Client) error {
	return npa.postCb(c, npa.nodeNameOrAddr, npa.extraVars)
}

func nodeCommission(c *manager.Client, nodeName, extraVars string) error {
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
		return errNodeNameMissing("maintenance")
	}
	return c.PostNodeInMaintenance(nodeName, extraVars)
}

func nodeDiscover(c *manager.Client, nodeAddr, extraVars string) error {
	if nodeAddr == "" {
		return errNodeAddrMissing("discover")
	}
	if ip := net.ParseIP(nodeAddr); ip == nil {
		return errInvalidIPAddr(nodeAddr)
	}
	return c.PostNodeDiscover(nodeAddr, extraVars)
}

func globalsSet(c *manager.Client, noop, extraVars string) error {
	return c.PostGlobals(extraVars)
}

type getActioner struct {
	nodeName string
	getCb    func(c *manager.Client, nodeName string) error
}

func newGetActioner(getCb func(c *manager.Client, nodeName string) error) *getActioner {
	return &getActioner{getCb: getCb}
}

func (nga *getActioner) procFlags(c *cli.Context) {
	return
}

func (nga *getActioner) procArgs(c *cli.Context) {
	nga.nodeName = c.Args().First()
}

func (nga *getActioner) action(c *manager.Client) error {
	return nga.getCb(c, nga.nodeName)
}

func nodeGet(c *manager.Client, nodeName string) error {
	if nodeName == "" {
		return errNodeNameMissing("get")
	}

	out, err := c.GetNode(nodeName)
	if err != nil {
		return err
	}

	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "    ")
	outBuf.WriteTo(os.Stdout)
	return nil
}

func nodesGet(c *manager.Client, noop string) error {
	out, err := c.GetAllNodes()
	if err != nil {
		return err
	}

	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "    ")
	outBuf.WriteTo(os.Stdout)
	return nil
}

func globalsGet(c *manager.Client, noop string) error {
	out, err := c.GetGlobals()
	if err != nil {
		return err
	}

	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "    ")
	outBuf.WriteTo(os.Stdout)
	return nil
}
