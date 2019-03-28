package umount

import (
	"log"

	"bazil.org/fuse"
	"github.com/ehotinger/lightningfs/defaults"
	"github.com/urfave/cli"
)

// Command performs an unmount.
var Command = cli.Command{
	Name:      "umount",
	Usage:     "perform an unmount",
	ArgsUsage: "[mount]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode",
		},
	},
	Action: func(context *cli.Context) error {
		var (
			mntPoint = context.Args().First()
			debug    = context.Bool("debug")
		)

		if mntPoint == "" {
			mntPoint = defaults.MntPoint
		}

		log.Printf("Attempting to unmount %s\n", mntPoint)

		if debug {
			fuse.Debug = func(msg interface{}) {
				log.Println(msg)
			}
		}

		return fuse.Unmount(mntPoint)
	},
}
