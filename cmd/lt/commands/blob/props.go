package blob

import (
	gocontext "context"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var propsCommand = cli.Command{
	Name:      "props",
	Usage:     "view blob properties",
	ArgsUsage: "",
	Flags: []cli.Flag{
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
			Name:  "blob-name",
			Usage: "The Azure Blob name",
		},
	},
	Action: func(context *cli.Context) error {
		var (
			accountName   = context.String("account-name")
			accountKey    = context.String("account-key")
			containerName = context.String("container-name")
			blobName      = context.String("blob-name")
		)

		if accountName == "" {
			return errors.New("account name is required")
		}
		if accountKey == "" {
			return errors.New("account key is required")
		}

		credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return errors.Wrap(err, "failed to create shared key credential")
		}

		p := azblob.NewPipeline(credential, azblob.PipelineOptions{
			Retry: azblob.RetryOptions{
				MaxTries:      1,
				MaxRetryDelay: 0,
				TryTimeout:    time.Second * 3,
			},
			// TODO: retries
		})

		cURL, err := url.Parse(fmt.Sprintf(blobFmt, accountName, containerName))
		if err != nil {
			return err
		}

		containerURL := azblob.NewContainerURL(*cURL, p)
		blockBlobURL := containerURL.NewBlockBlobURL(blobName)
		props, err := blockBlobURL.GetProperties(gocontext.Background(), azblob.BlobAccessConditions{})
		if err != nil {
			return err
		}
		fmt.Println(props.Date())
		return nil
	},
}
