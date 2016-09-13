package main

import (
	"github.com/Sirupsen/logrus"
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

	getJobFlags = []cli.Flag{
		jsonFlag,
		cli.BoolFlag{
			Name:  "follow, f",
			Usage: "stream job logs (just like tail -f). Only applicable for an active job",
		},
	}

	postFlags = []cli.Flag{
		extraVarsFlag,
	}

	postHostGroupFlags = []cli.Flag{
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
					Action:  doAction(newPostActioner(validateOneArg, nodeCommission)),
					Flags:   postHostGroupFlags,
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a node",
					Action:  doAction(newPostActioner(validateOneArg, nodeDecommission)),
					Flags:   postFlags,
				},
				{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "update a node",
					Action:  doAction(newPostActioner(validateOneArg, nodeUpdate)),
					Flags:   postHostGroupFlags,
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
					Flags:   postHostGroupFlags,
				},
				{
					Name:    "decommission",
					Aliases: []string{"d"},
					Usage:   "decommission a set of nodes",
					Action:  doAction(newPostActioner(validateMultiNodeNames, nodesDecommission)),
					Flags:   postFlags,
				},
				{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "update a set of nodes",
					Action:  doAction(newPostActioner(validateMultiNodeNames, nodesUpdate)),
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
					Flags:   getJobFlags,
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
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "set/get clusterm configuration",
			Subcommands: []cli.Command{
				{
					Name:    "get",
					Aliases: []string{"g"},
					Usage:   "get clusterm configuration",
					Action:  doAction(newGetActioner(configGet)),
					Flags:   getFlags,
				},
				{
					Name:    "set",
					Aliases: []string{"s"},
					Usage:   "set clusterm configuration. use '-' as the arg to read JSON configuration from stdin, else provide a path to the file containing JSON configuration",
					Action:  doAction(newPostActioner(validateOneArg, configSet)),
				},
			},
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
	streamLogs bool
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
			logrus.Fatalf(err.Error())
		}
	}
}
