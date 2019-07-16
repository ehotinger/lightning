package mount

import (
	"fmt"
	"log"
	"os"
	"runtime"

	gocontext "context"

	"github.com/ehotinger/lightningfs/config"
	"github.com/ehotinger/lightningfs/defaults"
	"github.com/ehotinger/lightningfs/fs"
	"github.com/jacobsa/fuse"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Command performs a mount.
var Command = cli.Command{
	Name:      "mount",
	Usage:     "perform a mount",
	ArgsUsage: "[mount]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode",
		},
		cli.StringFlag{
			Name:  "account-name",
			Usage: "Azure Blob account name",
		},
		cli.StringFlag{
			Name:  "account-key",
			Usage: "Azure Blob account key",
		},
		cli.StringFlag{
			Name:  "container-name",
			Usage: "Azure Blob container name",
		},
		cli.StringFlag{
			Name:  "cache-path",
			Usage: "The location of the disk cache",
		},
		cli.StringFlag{
			Name:  "config-file",
			Usage: "The location of the configuration file",
		},
	},
	Action: func(context *cli.Context) error {
		var (
			mntPoint   = context.Args().First()
			debug      = context.Bool("debug")
			configFile = context.String("config-file")
		)

		var cfg *config.Config
		if configFile == "" {
			accountName := context.String("account-name")
			accountKey := context.String("account-key")
			containerName := context.String("container-name")
			cachePath := context.String("cache-path")
			cfg = config.NewConfig(accountName, accountKey, containerName, cachePath)
		} else {
			log.Println("Loading configuration...")
			var err error
			cfg, err = config.NewConfigFromFile(configFile)
			if err != nil {
				return err
			}
		}

		if mntPoint == "" {
			mntPoint = defaults.MntPoint
		}

		fmt.Fprintf(os.Stdout, "Using %s as the mount point\n", mntPoint)

		if cfg.AzureAccountName == "" {
			return errors.New("account name is required")
		}
		if cfg.AzureAccountKey == "" {
			return errors.New("account key is required")
		}

		// TODO:
		// Allow parallelism in the file system implementation
		// to help flush out potential bugs.
		runtime.GOMAXPROCS(2)
		server, err := fs.NewLightningFS(cfg, 0, 0)
		if err != nil {
			log.Fatalf("failed to setup server: %v", err)
		}
		fuseCfg := &fuse.MountConfig{
			ReadOnly: false,
			FSName:   "lightningfs",
		}
		if debug {
			fuseCfg.DebugLogger = log.New(os.Stdout, "DEBUG: ", 0)
		}

		mountedFS, err := fuse.Mount(mntPoint, server, fuseCfg)
		if err != nil {
			log.Fatalf("failed to mount: %v", err)
		}

		if err = mountedFS.Join(gocontext.Background()); err != nil {
			log.Fatalf("failed to unmount: %v", err)
		}

		return nil
	},
}
