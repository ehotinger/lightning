package fs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/ehotinger/lightningfs/config"
	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
	"github.com/pkg/errors"
)

const (
	blobFmt = "https://%s.blob.core.windows.net/%s"
)

func NewLightningFS(config *config.Config, uid uint32, gid uint32) (server fuse.Server, err error) {
	// TODO: SAS support
	credential, err := azblob.NewSharedKeyCredential(config.AzureAccountName, config.AzureAccountKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared key credential")
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{}, // TODO: retries
	})

	cURL, err := url.Parse(fmt.Sprintf(blobFmt, config.AzureAccountName, config.ContainerName))
	if err != nil {
		return nil, err
	}

	containerURL := azblob.NewContainerURL(*cURL, p)

	fs := &lightningFS{
		containerURL: containerURL,
		inodes:       make([]*iNode, fuseops.RootInodeID+1),
		uid:          uid,
		gid:          gid,
	}

	now := time.Now()
	// Set up the root
	fs.inodes[fuseops.RootInodeID] = newINode(
		fuseops.InodeAttributes{
			Mode:   0700 | os.ModeDir,
			Uid:    uid,
			Gid:    gid,
			Mtime:  now,
			Crtime: now,
			Atime:  now,
		},
	)

	server = fuseutil.NewFileSystemServer(fs)
	return server, nil
}

type lightningFS struct {
	containerURL azblob.ContainerURL

	mu     sync.RWMutex
	inodes []*iNode
	uid    uint32
	gid    uint32
}

// Statfs obtains the file system's metadata.
func (fs *lightningFS) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) (err error) {
	// TODO: modify the return?
	return
}

func (fs *lightningFS) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	parent, err := fs.getINode(op.Parent)
	if err != nil {
		return err
	}

	childID, _, ok := parent.LookUpChild(op.Name)
	if !ok {
		return fuse.ENOENT
	}

	child, err := fs.getINode(childID)
	if err != nil {
		return err
	}

	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs
	op.Entry.AttributesExpiration = getDefaultAttributesExpiration()
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration
	return nil
}

func (fs *lightningFS) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	inode, err := fs.getINode(op.Inode)
	if err != nil {
		return err
	}

	op.Attributes = inode.attrs
	op.AttributesExpiration = getDefaultAttributesExpiration()
	return nil
}

func (fs *lightningFS) SetInodeAttributes(
	ctx context.Context,
	op *fuseops.SetInodeAttributesOp) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	inode, err := fs.getINode(op.Inode)
	if err != nil {
		return err
	}

	inode.setAttributes(op.Size, op.Mode, op.Mtime)
	op.Attributes = inode.attrs
	op.AttributesExpiration = getDefaultAttributesExpiration()
	return nil
}

func (fs *lightningFS) ForgetInode(
	ctx context.Context,
	op *fuseops.ForgetInodeOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) MkDir(
	ctx context.Context,
	op *fuseops.MkDirOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) MkNode(
	ctx context.Context,
	op *fuseops.MkNodeOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) CreateFile(
	ctx context.Context,
	op *fuseops.CreateFileOp) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var err error
	op.Entry, err = fs.createFile(op.Parent, op.Name, op.Mode)
	return err
}

func (fs *lightningFS) CreateSymlink(
	ctx context.Context,
	op *fuseops.CreateSymlinkOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) CreateLink(
	ctx context.Context,
	op *fuseops.CreateLinkOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) Rename(
	ctx context.Context,
	op *fuseops.RenameOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) RmDir(
	ctx context.Context,
	op *fuseops.RmDirOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) Unlink(
	ctx context.Context,
	op *fuseops.UnlinkOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	inode, err := fs.getINode(op.Inode)
	if err != nil {
		return err
	}

	if !inode.isDir() {
		return errors.New("node is not a directory")
	}

	return nil
}

func (fs *lightningFS) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) (err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	inode, err := fs.getINode(op.Inode)
	if err != nil {
		return err
	}

	op.BytesRead = inode.readDir(op.Dst, int(op.Offset))
	return
}

func (fs *lightningFS) ReleaseDirHandle(
	ctx context.Context,
	op *fuseops.ReleaseDirHandleOp) error {
	return fuse.ENOSYS // TODO: Unimplemented
}

// NB: errors are ignored by the kernel
func (fs *lightningFS) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}
func (fs *lightningFS) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) (err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	inode, err := fs.getINode(op.Inode)
	if err != nil {
		return err
	}
	op.BytesRead, err = inode.readAt(op.Dst, op.Offset)

	// Don't return EOF errors; we just indicate EOF to fuse using a short read.
	if err == io.EOF {
		err = nil
	}

	return
}

func (fs *lightningFS) WriteFile(
	ctx context.Context,
	op *fuseops.WriteFileOp) (err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	inode, err := fs.getINode(op.Inode)
	if err != nil {
		return err
	}
	_, err = inode.writeAt(op.Data, op.Offset)
	return
}

func (fs *lightningFS) SyncFile(
	ctx context.Context,
	op *fuseops.SyncFileOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) FlushFile(
	ctx context.Context,
	op *fuseops.FlushFileOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) ReleaseFileHandle(
	ctx context.Context,
	op *fuseops.ReleaseFileHandleOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) ReadSymlink(
	ctx context.Context,
	op *fuseops.ReadSymlinkOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) RemoveXattr(
	ctx context.Context,
	op *fuseops.RemoveXattrOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) GetXattr(
	ctx context.Context,
	op *fuseops.GetXattrOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) ListXattr(
	ctx context.Context,
	op *fuseops.ListXattrOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) SetXattr(
	ctx context.Context,
	op *fuseops.SetXattrOp) (err error) {
	return fuse.ENOSYS // TODO: Unimplemented
}

func (fs *lightningFS) Destroy() {}
