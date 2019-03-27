package main

import (
	"fmt"
	"os"

	mountCmd "github.com/ehotinger/lightning/cmd/lt/commands/mount"
	versionCmd "github.com/ehotinger/lightning/cmd/lt/commands/version"
	"github.com/ehotinger/lightning/version"
	"github.com/urfave/cli"
)

func main() {
	app := New()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// New returns a *cli.App instance.
func New() *cli.App {
	app := cli.NewApp()
	app.Name = "lt"
	app.Usage = "Mount Azure blobs at lightning speed"
	app.Version = version.Version
	app.Commands = []cli.Command{
		mountCmd.Command,
		versionCmd.Command,
	}
	return app
}
