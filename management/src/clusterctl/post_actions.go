package main

import (
	"net"

	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
)

type postCallback func(c *manager.Client, args []string, flags parsedFlags) error
type validateCallback func(args []string) error

type parsedFlags struct {
	extraVars string
	hostGroup string
}

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
