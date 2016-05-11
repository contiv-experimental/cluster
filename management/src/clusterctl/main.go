package main

import (
	"os"

	"github.com/codegangsta/cli"
)

// version is provided by build
var version = ""

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Version = version
	app.Usage = "utility to interact with cluster manager"
	app.Flags = clustermFlags
	app.Commands = commands
	app.Run(os.Args)
}
