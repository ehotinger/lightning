package blob

import (
	"github.com/urfave/cli"
)

const (
	blobFmt = "https://%s.blob.core.windows.net/%s"
)

// Command performs blob operations.
var Command = cli.Command{
	Name:      "blob",
	Usage:     "various blob commands",
	ArgsUsage: "",
	Flags:     []cli.Flag{},
	Subcommands: []cli.Command{
		propsCommand,
		uploadCommand,
	},
}
