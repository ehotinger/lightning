package main

import (
	"fmt"
	"os"

	blobCmd "github.com/ehotinger/lightningfs/cmd/lt/commands/blob"
	mountCmd "github.com/ehotinger/lightningfs/cmd/lt/commands/mount"
	umountCmd "github.com/ehotinger/lightningfs/cmd/lt/commands/umount"
	versionCmd "github.com/ehotinger/lightningfs/cmd/lt/commands/version"
	"github.com/ehotinger/lightningfs/version"
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
		blobCmd.Command,
		mountCmd.Command,
		umountCmd.Command,
		versionCmd.Command,
	}
	return app
}
