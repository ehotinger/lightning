package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	fuseFS "bazil.org/fuse/fs"
	lightningFS "github.com/ehotinger/lightning/fs"
)

const (
	defaultMntPoint = "/mnt/lightning"
)

func main() {
	mntPoint := defaultMntPoint
	flag.Parse()
	if flag.NArg() > 1 {
		printUsage()
		os.Exit(2)
	} else if flag.NArg() == 1 {
		mntPoint = flag.Arg(0)
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
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <MOUNT_POINT> (defaults to %s)\n", os.Args[0], defaultMntPoint)
	flag.PrintDefaults()
}
