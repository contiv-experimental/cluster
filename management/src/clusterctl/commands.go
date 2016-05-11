package main

import (
	"log"

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

	jsonFlag = cli.BoolFlag{
		Name:  "json, j",
		Usage: "print command output in JSON",
	}

	getFlags = []cli.Flag{
		jsonFlag,
	}

	postFlags = []cli.Flag{
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

	commands = []cli.Command{
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
					Flags:   postFlags,
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a node in maintenance",
					Action:  doAction(newPostActioner(validateOneNodeName, nodeMaintenance)),
					Flags:   postFlags,
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get node's status information",
					Action:  doAction(newGetActioner(nodeGet)),
					Flags:   getFlags,
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
					Flags:   postFlags,
				},
				{
					Name:    "maintenance",
					Aliases: []string{"m"},
					Usage:   "put a set of nodes in maintenance",
					Action:  doAction(newPostActioner(validateMultiNodeNames, nodesMaintenance)),
					Flags:   postFlags,
				},
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get status information for all nodes",
					Action:  doAction(newGetActioner(nodesGet)),
					Flags:   getFlags,
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
					Flags:   getFlags,
				},
				{
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "set global info",
					Action:  doAction(newPostActioner(validateZeroArgs, globalsSet)),
					Flags:   postFlags,
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
					Flags:   getFlags,
				},
			},
		},
		{
			Name:    "discover",
			Aliases: []string{"d"},
			Usage:   "provision one or more nodes for discovery",
			Action:  doAction(newPostActioner(validateMultiNodeAddrs, nodesDiscover)),
			Flags:   postFlags,
		},
	}
)

func errUnexpectedArgCount(exptd string, rcvd int) error {
	return errored.Errorf("command expects %s arg(s) but received %d", exptd, rcvd)
}

func errInvalidIPAddr(a string) error {
	return errored.Errorf("failed to parse ip address %q", a)
}

type parsedFlags struct {
	extraVars  string
	hostGroup  string
	jsonOutput bool
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
