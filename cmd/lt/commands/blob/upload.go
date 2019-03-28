package blob

import (
	gocontext "context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// TODO: centralize flags
var uploadCommand = cli.Command{
	Name:      "upload",
	Usage:     "upload a blob",
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

		requestBody := strings.NewReader("some text")
		_, err = blockBlobURL.Upload(gocontext.Background(),
			pipeline.NewRequestBodyProgress(requestBody, func(bytesTransferred int64) {
				fmt.Printf("Wrote %d of %d bytes.", bytesTransferred, requestBody.Size())
			}),
			azblob.BlobHTTPHeaders{
				ContentType:        "text/html; charset=utf-8",
				ContentDisposition: "attachment",
			}, azblob.Metadata{}, azblob.BlobAccessConditions{})
		return err
	},
}
