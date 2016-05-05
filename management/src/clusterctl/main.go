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
	clustermFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "url, u",
			Value: manager.DefaultConfig().Manager.Addr,
			Usage: "cluster manager's REST service url",
		},
	}

	extraVarsFlag = cli.StringFlag{
		Name:  "extra-vars, e",
		Value: "",
		Usage: "extra vars for ansible configuration. This should be a quoted json string.",
	}

	cmdFlags = []cli.Flag{
		extraVarsFlag,
	}

	commissionFlags = []cli.Flag{
		extraVarsFlag,
		cli.StringFlag{
			Name:  "host-group, g",
			Value: "",
			Usage: "host-group of the node(s). Possible values: service-master or service-worker",
		},
	}
)

// parsedFlags is used to save cmdFlags
type parsedFlags struct {
	extraVars string
	hostGroup string
}

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
					Action:  doAction(newPostActioner(validateOneNodeName, nodeCommission)),
					Flags:   commissionFlags,
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a node",
					Action:  doAction(newPostActioner(validateOneNodeName, nodeDecommission)),
					Flags:   cmdFlags,
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a node in maintenance",
					Action:  doAction(newPostActioner(validateOneNodeName, nodeMaintenance)),
					Flags:   cmdFlags,
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
					Name:    "commission",
					Aliases: []string{"c"},
					Usage:   "commission a set of nodes",
					Action:  doAction(newPostActioner(validateMultiNodeNames, nodesCommission)),
					Flags:   commissionFlags,
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a set of nodes",
					Action:  doAction(newPostActioner(validateMultiNodeNames, nodesDecommission)),
					Flags:   cmdFlags,
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a set of nodes in maintenance",
					Action:  doAction(newPostActioner(validateMultiNodeNames, nodesMaintenance)),
					Flags:   cmdFlags,
				},
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
					Flags:   cmdFlags,
					Action:  doAction(newPostActioner(validateZeroArgs, globalsSet)),
				},
			},
		},
		{
			Name:    "job",
			Aliases: []string{"j"},
			Usage:   "provisioning job info",
			Subcommands: []cli.Command{
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get job info. Expects an arg with value 'active' or 'last'",
					Action:  doAction(newGetActioner(jobGet)),
				},
			},
		},
		{
			Name:    "discover",
			Aliases: []string{"d"},
			Usage:   "provision one or more nodes for discovery",
			Action:  doAction(newPostActioner(validateMultiNodeAddrs, nodesDiscover)),
			Flags:   cmdFlags,
		},
	}

	app.Run(os.Args)
}

func errUnexpectedArgCount(exptd string, rcvd int) error {
	return errored.Errorf("command expects %s arg(s) but received %d", exptd, rcvd)
}

func errInvalidIPAddr(a string) error {
	return errored.Errorf("failed to parse ip address %q", a)
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

type postCallback func(c *manager.Client, args []string, flags parsedFlags) error
type validateCallback func(args []string) error

type postActioner struct {
	args       []string
	flags      parsedFlags
	validateCb validateCallback
	postCb     postCallback
}

func newPostActioner(validateCb validateCallback, postCb postCallback) *postActioner {
	return &postActioner{
		validateCb: validateCb,
		postCb:     postCb,
	}
}

func (npa *postActioner) procFlags(c *cli.Context) {
	npa.flags.extraVars = c.String("extra-vars")
	npa.flags.hostGroup = c.String("host-group")
}

func (npa *postActioner) procArgs(c *cli.Context) {
	npa.args = c.Args()
}

func (npa *postActioner) action(c *manager.Client) error {
	if err := npa.validateCb(npa.args); err != nil {
		return err
	}
	return npa.postCb(c, npa.args, npa.flags)
}

func validateOneNodeName(args []string) error {
	if len(args) != 1 {
		return errUnexpectedArgCount("1", len(args))
	}
	return nil
}

func nodeCommission(c *manager.Client, args []string, flags parsedFlags) error {
	nodeName := args[0]
	return c.PostNodeCommission(nodeName, flags.extraVars, flags.hostGroup)
}

func nodeDecommission(c *manager.Client, args []string, flags parsedFlags) error {
	nodeName := args[0]
	return c.PostNodeDecommission(nodeName, flags.extraVars)
}

func nodeMaintenance(c *manager.Client, args []string, flags parsedFlags) error {
	nodeName := args[0]
	return c.PostNodeInMaintenance(nodeName, flags.extraVars)
}

func validateMultiNodeNames(args []string) error {
	if len(args) < 1 {
		return errUnexpectedArgCount(">=1", len(args))
	}
	return nil
}

func nodesCommission(c *manager.Client, args []string, flags parsedFlags) error {
	return c.PostNodesCommission(args, flags.extraVars, flags.hostGroup)
}

func nodesDecommission(c *manager.Client, args []string, flags parsedFlags) error {
	return c.PostNodesDecommission(args, flags.extraVars)
}

func nodesMaintenance(c *manager.Client, args []string, flags parsedFlags) error {
	return c.PostNodesInMaintenance(args, flags.extraVars)
}

func validateMultiNodeAddrs(args []string) error {
	if len(args) < 1 {
		return errUnexpectedArgCount(">=1", len(args))
	}
	for _, addr := range args {
		if ip := net.ParseIP(addr); ip == nil {
			return errInvalidIPAddr(addr)
		}
	}
	return nil
}

func nodesDiscover(c *manager.Client, args []string, flags parsedFlags) error {
	return c.PostNodesDiscover(args, flags.extraVars)
}

func validateZeroArgs(args []string) error {
	if len(args) != 0 {
		return errUnexpectedArgCount("0", len(args))
	}
	return nil
}

func globalsSet(c *manager.Client, noop []string, flags parsedFlags) error {
	return c.PostGlobals(flags.extraVars)
}

type getActioner struct {
	arg   string
	getCb func(c *manager.Client, arg string) error
}

func newGetActioner(getCb func(c *manager.Client, nodeName string) error) *getActioner {
	return &getActioner{getCb: getCb}
}

func (nga *getActioner) procFlags(c *cli.Context) {
	return
}

func (nga *getActioner) procArgs(c *cli.Context) {
	nga.arg = c.Args().First()
}

func (nga *getActioner) action(c *manager.Client) error {
	return nga.getCb(c, nga.arg)
}

func ppJSON(out []byte) {
	var outBuf bytes.Buffer
	json.Indent(&outBuf, out, "", "    ")
	outBuf.WriteTo(os.Stdout)
}

func nodeGet(c *manager.Client, nodeName string) error {
	if nodeName == "" {
		return errUnexpectedArgCount("1", 0)
	}

	out, err := c.GetNode(nodeName)
	if err != nil {
		return err
	}

	ppJSON(out)
	return nil
}

func nodesGet(c *manager.Client, noop string) error {
	out, err := c.GetAllNodes()
	if err != nil {
		return err
	}

	ppJSON(out)
	return nil
}

func globalsGet(c *manager.Client, noop string) error {
	out, err := c.GetGlobals()
	if err != nil {
		return err
	}

	ppJSON(out)
	return nil
}

func jobGet(c *manager.Client, job string) error {
	if job == "" {
		return errUnexpectedArgCount("1", 0)
	}

	out, err := c.GetJob(job)
	if err != nil {
		return err
	}

	ppJSON(out)
	return nil
}
