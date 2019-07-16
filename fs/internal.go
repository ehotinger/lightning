package fs

import (
	"fmt"
	"os"
	"time"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

// getINode returns an iNode if it's allocated and returns an error otherwise.
func (fs *lightningFS) getINode(id fuseops.InodeID) (*iNode, error) {
	numINodes := fuseops.InodeID(len(fs.inodes))
	if id > numINodes {
		return nil, fmt.Errorf("id: %v out of range (max: %v)", id, numINodes)
	}
	inode := fs.inodes[id]
	if inode == nil {
		return nil, fmt.Errorf("unknown inode: %v", id)
	}
	return inode, nil
}

// getDefaultAttributesExpiration returns a default attributes expiration time.
// We allow the kernel to cache as long as it wants to and handle invalidation.
func getDefaultAttributesExpiration() time.Time {
	return time.Now().Add(365 * 24 * time.Hour)
}

func (fs *lightningFS) createFile(
	parentID fuseops.InodeID,
	name string,
	mode os.FileMode) (entry fuseops.ChildInodeEntry, err error) {

	parent, err := fs.getINode(parentID)
	if err != nil {
		return entry, err
	}

	// Don't create a duplicate
	_, _, exists := parent.LookUpChild(name)
	if exists {
		err = fuse.EEXIST
		return
	}

	now := time.Now()
	childAttrs := fuseops.InodeAttributes{
		Nlink:  1,
		Mode:   mode,
		Atime:  now,
		Mtime:  now,
		Ctime:  now,
		Crtime: now,
		Uid:    fs.uid,
		Gid:    fs.gid,
	}

	childID, child := fs.allocateInode(childAttrs)
	parent.addChild(childID, name, fuseutil.DT_File)

	entry.Child = childID
	entry.Attributes = child.attrs
	entry.AttributesExpiration = getDefaultAttributesExpiration()
	entry.EntryExpiration = entry.AttributesExpiration
	return
}

func (fs *lightningFS) allocateInode(attrs fuseops.InodeAttributes) (id fuseops.InodeID, inode *iNode) {
	inode = newINode(attrs)
	// TODO: re-use a free ID if possible, otherwise create a new one.
	id = fuseops.InodeID(len(fs.inodes))
	fs.inodes = append(fs.inodes, inode)
	return
}
