package main

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/codegangsta/cli"
	"github.com/contiv/cluster/management/src/clusterm/manager"
)

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
