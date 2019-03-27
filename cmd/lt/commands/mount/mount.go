package mount

import (
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	fuseFS "bazil.org/fuse/fs"
	lightningFS "github.com/ehotinger/lightningfs/fs"
	"github.com/urfave/cli"
)

const (
	defaultMntPoint = "/mnt/lightning"
)

// Command performs a mount.
var Command = cli.Command{
	Name:  "mount",
	Usage: "perform a mount",
	Action: func(context *cli.Context) error {
		var (
			mntPoint = context.Args().First()
			debug    = context.Bool("debug")
		)
		if mntPoint == "" {
			mntPoint = defaultMntPoint
		}

		if debug {
			fuse.Debug = func(msg interface{}) {
				log.Println("[]", msg)
			}
		}

		fmt.Fprintf(os.Stdout, "Using %s as the mount point\n", mntPoint)

		c, err := fuse.Mount(mntPoint, fuse.FSName("ltfs"), fuse.Subtype("ltfs"), fuse.ReadOnly())
		if err != nil {
			log.Fatalf("failed to perform fuse mount, err: %v", err)
		}
		defer c.Close()
		defer fuse.Unmount(mntPoint)

		ltFS, err := lightningFS.NewLightningFS()
		if err != nil {
			log.Fatal(err)
		}

		err = fuseFS.Serve(c, ltFS)
		if err != nil {
			log.Fatal(err)
		}

		<-c.Ready
		if err := c.MountError; err != nil {
			log.Fatal(err)
		}

		return nil
	},
}
