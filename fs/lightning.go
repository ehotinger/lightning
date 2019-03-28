package fs

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"bazil.org/fuse"
	fuseFS "bazil.org/fuse/fs"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/ehotinger/lightningfs/config"
	"github.com/pkg/errors"
)

const (
	blobFmt = "https://%s.blob.core.windows.net/%s"
	debug   = true // TODO: remove
)

type LightningFS struct {
	containerURL azblob.ContainerURL
}

func NewLightningFS(config *config.Config) (*LightningFS, error) {
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

	return &LightningFS{
		containerURL: containerURL,
	}, nil
}

//
// FS interfaces
//

// Root is called to obtain the Node for the file system root.
func (fs *LightningFS) Root() (fuseFS.Node, error) {
	return &DirBlob{
		containerURL: fs.containerURL,
	}, nil
}

type DirBlob struct {
	name         string
	containerURL azblob.ContainerURL
	blockBlobURL azblob.BlockBlobURL
}

type FileBlob struct {
	// name         string
	// blockBlobURL azblob.BlockBlobURL
}

// Statfs is called to obtain file system metadata.
// It should write that data to resp.
func (fs *LightningFS) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
	//type FSStatfser interface {
	return nil
}

// Destroy is called when the file system is shutting down.
//
// Linux only sends this request for block device backed (fuseblk)
// filesystems, to allow them to flush writes to disk before the
// unmount completes.
func (fs *LightningFS) Destroy() {}

// GenerateInode is called to pick a dynamic inode number when it
// would otherwise be 0.
//
// Not all filesystems bother tracking inodes, but FUSE requires
// the inode to be set, and fewer duplicates in general makes UNIX
// tools work better.
//
// Operations where the nodes may return 0 inodes include Getattr,
// Setattr and ReadDir.
//
// If FS does not implement FSInodeGenerator, GenerateDynamicInode
// is used.
//
// Implementing this is useful to e.g. constrain the range of
// inode values used for dynamic inodes.
func (fs *LightningFS) GenerateInode(parentInode uint64, name string) uint64 {
	// type FSInodeGenerator interface {
	return 0
}

//
// Node interfaces
//

// A Blob represents a node in the graph.

// A Node is the interface required of a file or directory.
// See the documentation for type FS for general information
// pertaining to all methods.
//
// A Node must be usable as a map key, that is, it cannot be a
// function, map or slice.
//
// Other FUSE requests can be handled by implementing methods from the
// Node* interfaces, for example NodeOpener.
//
// Methods returning Node should take care to return the same Node
// when the result is logically the same instance. Without this, each
// Node will get a new NodeID, causing spurious cache invalidations,
// extra lookups and aliasing anomalies. This may not matter for a
// simple, read-only filesystem.

func getMetadataWithDefaults(ctx context.Context, blockBlobURL azblob.BlockBlobURL) (azblob.Metadata, error) {
	props, err := blockBlobURL.GetProperties(ctx, azblob.BlobAccessConditions{})
	if err != nil {
		return nil, err
	}

	metadata := props.NewMetadata()
	tryDefault(metadata, "size", "0")
	// TODO:
	// Valid time.Duration // how long Attr can be cached
	// Inode     uint64      // inode number
	// Size      uint64      // size in bytes
	// Blocks    uint64      // size in 512-byte units
	// Atime     time.Time   // time of last access
	// Mtime     time.Time   // time of last modification
	// Ctime     time.Time   // time of last inode change
	// Crtime    time.Time   // time of creation (OS X only)
	// Mode      os.FileMode // file mode
	// Nlink     uint32      // number of links (usually 1)
	// Uid       uint32      // owner uid
	// Gid       uint32      // group gid
	// Rdev      uint32      // device numbers
	// Flags     uint32      // chflags(2) flags (OS X only)
	// BlockSize uint32      // preferred blocksize for filesystem I/O
	return metadata, nil
}

// tryDefault defaults the specified property's value if it doesn't exist already.
func tryDefault(metadata azblob.Metadata, property string, value string) {
	if _, ok := metadata[property]; !ok {
		metadata[property] = value
	}
}

func blobAttr(ctx context.Context, a *fuse.Attr, blockBlobURL azblob.BlockBlobURL) error {
	metadata, err := getMetadataWithDefaults(ctx, blockBlobURL)
	if err != nil {
		return err
	}

	size, err := strconv.ParseUint(metadata["size"], 10, 64)
	if err != nil {
		return err
	}
	a.Size = size

	if debug {
		log.Printf("attr of %v: %s\n", blockBlobURL.URL(), *a)
	}

	return nil
}

// Attr fills attr with the standard metadata for the node.
//
// Fields with reasonable defaults are prepopulated. For example,
// all times are set to a fixed moment when the program started.
//
// If Inode is left as 0, a dynamic inode number is chosen.
//
// The result may be cached for the duration set in Valid.
func (b *DirBlob) Attr(ctx context.Context, a *fuse.Attr) error {
	// Root directory
	if b.name == "" {
		a.Mode = os.ModeDir | 0755
		return nil
	}

	return blobAttr(ctx, a, b.blockBlobURL)
}

// Setattr sets the standard metadata for the receiver.
//
// Note, this is also used to communicate changes in the size of
// the file, outside of Writes.
//
// req.Valid is a bitmask of what fields are actually being set.
// For example, the method should not change the mode of the file
// unless req.Valid.Mode() is true.

func (b *DirBlob) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// type NodeSetattrer interface {
	log.Println("DirBlob: SetAttr")
	return nil
}

// Symlink creates a new symbolic link in the receiver, which must be a directory.
//
// TODO is the above true about directories?
func (b *DirBlob) Symlink(ctx context.Context, req *fuse.SymlinkRequest) (fuseFS.Node, error) {
	// type NodeSymlinker interface {
	log.Println("DirBlob: Symlink")
	return nil, nil
}

// Readlink reads a symbolic link.
// This optional request will be called only for symbolic link nodes.
func (b *DirBlob) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	// type NodeReadlinker interface {
	log.Println("DirBlob: Readlink")
	return "", nil
}

// Link creates a new directory entry in the receiver based on an
// existing Node. Receiver must be a directory.
func (b *DirBlob) Link(ctx context.Context, req *fuse.LinkRequest, old fuseFS.Node) (fuseFS.Node, error) {
	// type NodeLinker interface {
	log.Println("DirBlob: Link")
	return nil, nil
}

// Remove removes the entry with the given name from
// the receiver, which must be a directory.  The entry to be removed
// may correspond to a file (unlink) or to a directory (rmdir).
func (b *DirBlob) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	// type NodeRemover interface {
	log.Println("DirBlob: Remove")
	return nil
}

// Access checks whether the calling context has permission for
// the given operations on the receiver. If so, Access should
// return nil. If not, Access should return EPERM.
//
// Note that this call affects the result of the access(2) system
// call but not the open(2) system call. If Access is not
// implemented, the Node behaves as if it always returns nil
// (permission granted), relying on checks in Open instead.
func (b *DirBlob) Access(ctx context.Context, req *fuse.AccessRequest) error {
	log.Println("DirBlob: Access")
	return nil
}

// Lookup looks up a specific entry in the receiver,
// which must be a directory.  Lookup should return a Node
// corresponding to the entry.  If the name does not exist in
// the directory, Lookup should return ENOENT.
//
// Lookup need not to handle the names "." and "..".
func (b *DirBlob) Lookup(ctx context.Context, name string) (fuseFS.Node, error) {
	// type NodeStringLookuper interface {
	log.Println("DirBlob: Lookup")
	return nil, nil
}

// TODO: see  NodeRequestLookuper
// func (b *Blob) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fuseFS.Node, error) {
// 	return nil, nil
// }

// Open opens the receiver. After a successful open, a client
// process has a file descriptor referring to this Handle.
//
// Open can also be also called on non-files. For example,
// directories are Opened for ReadDir or fchdir(2).
//
// If this method is not implemented, the open will always
// succeed, and the Node itself will be used as the Handle.
//
// XXX note about access.  XXX OpenFlags.
func (b *DirBlob) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fuseFS.Handle, error) {
	// type NodeOpener interface {
	log.Println("DirBlob: Open")
	return nil, nil
}

// Create creates a new directory entry in the receiver, which
// must be a directory.
func (b *DirBlob) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fuseFS.Node, fuseFS.Handle, error) {
	// type NodeCreater interface {
	log.Println("DirBlob: Create")
	return nil, nil, nil
}

// Forget about this node. This node will not receive further
// method calls.
//
// Forget is not necessarily seen on unmount, as all nodes are
// implicitly forgotten as part part of the unmount.
func (b *DirBlob) Forget() {
	log.Println("DirBlob: Forget")
	// type NodeForgetter interface {
}

func (b *DirBlob) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fuseFS.Node) error {
	// NodeRenamer
	log.Println("DirBlob: Rename")
	return nil
}

func (b *DirBlob) Mknod(ctx context.Context, req *fuse.MknodRequest) (fuseFS.Node, error) {
	// NodeMknoder
	log.Println("DirBlob: Mknod")
	return nil, nil
}

// TODO this should be on Handle not Node
func (b *DirBlob) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Println("DirBlob: Fsync")
	return nil
}

// Getxattr gets an extended attribute by the given name from the
// node.
//
// If there is no xattr by that name, returns fuse.ErrNoXattr.
func (b *DirBlob) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	// type NodeGetxattrer interface {
	log.Println("DirBlob: Getxattr")
	return nil
}

// Listxattr lists the extended attributes recorded for the node.
func (b *DirBlob) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	// NodeListxattrer
	log.Println("DirBlob: Listxattr")
	return nil
}

// Setxattr sets an extended attribute with the given name and
// value for the node.
func (b *DirBlob) Setxattr(ctx context.Context, req *fuse.SetxattrRequest) error {
	log.Println("DirBlob: Setxattr")
	return nil
}

// Removexattr removes an extended attribute for the name.
//
// If there is no xattr by that name, returns fuse.ErrNoXattr.
func (b *DirBlob) Removexattr(ctx context.Context, req *fuse.RemovexattrRequest) error {
	log.Println("DirBlob: Removexattr")
	return nil
}

// A Handle is the interface required of an opened file or directory.
// See the documentation for type FS for general information
// pertaining to all methods.
//
// Other FUSE requests can be handled by implementing methods from the
// Handle* interfaces. The most common to implement are HandleReader,
// HandleReadDirer, and HandleWriter.
//
// TODO implement methods: Getlk, Setlk, Setlkw
type FileHandle struct{}

// type HandleFlusher interface {
// 	// Flush is called each time the file or directory is closed.
// 	// Because there can be multiple file descriptors referring to a
// 	// single opened file, Flush can be called multiple times.
// 	Flush(ctx context.Context, req *fuse.FlushRequest) error
// }

// type HandleReadAller interface {
// 	ReadAll(ctx context.Context) ([]byte, error)
// }

// type HandleReadDirAller interface {
// 	ReadDirAll(ctx context.Context) ([]fuse.Dirent, error)
// }

// Read requests to read data from the handle.
//
// There is a page cache in the kernel that normally submits only
// page-aligned reads spanning one or more pages. However, you
// should not rely on this. To see individual requests as
// submitted by the file system clients, set OpenDirectIO.
//
// Note that reads beyond the size of the file as reported by Attr
// are not even attempted (except in OpenDirectIO mode).
func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("fh: Read")
	return nil
}

// Write requests to write data into the handle at the given offset.
// Store the amount of data written in resp.Size.
//
// There is a writeback page cache in the kernel that normally submits
// only page-aligned writes spanning one or more pages. However,
// you should not rely on this. To see individual requests as
// submitted by the file system clients, set OpenDirectIO.
//
// Writes that grow the file are expected to update the file size
// (as seen through Attr). Note that file size changes are
// communicated also through Setattr.
func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("fh: Write")
	return nil
}

// type HandleReleaser interface {
// 	Release(ctx context.Context, req *fuse.ReleaseRequest) error
// }

// type Config struct {
// 	// Function to send debug log messages to. If nil, use fuse.Debug.
// 	// Note that changing this or fuse.Debug may not affect existing
// 	// calls to Serve.
// 	//
// 	// See fuse.Debug for the rules that log functions must follow.
// 	Debug func(msg interface{})

// 	// Function to put things into context for processing the request.
// 	// The returned context must have ctx as its parent.
// 	//
// 	// Note that changing this may not affect existing calls to Serve.
// 	//
// 	// Must not retain req.
// 	WithContext func(ctx context.Context, req fuse.Request) context.Context
// }
