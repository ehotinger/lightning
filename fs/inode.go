package fs

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

type iNode struct {
	attrs    fuseops.InodeAttributes
	entries  []fuseutil.Dirent
	contents []byte
	xattrs   map[string][]byte
}

func (in *iNode) equals(other *iNode) bool {
	if in == nil && other == nil {
		return true
	}
	if in == nil || other == nil {
		return false
	}
	return reflect.DeepEqual(in.attrs, other.attrs) &&
		reflect.DeepEqual(in.entries, other.entries) &&
		reflect.DeepEqual(in.contents, other.contents) &&
		reflect.DeepEqual(in.xattrs, other.xattrs)
}

func newINode(attrs fuseops.InodeAttributes) (in *iNode) {
	in = &iNode{
		attrs:  attrs,
		xattrs: make(map[string][]byte),
	}
	return
}

func (in *iNode) LookUpChild(name string) (
	id fuseops.InodeID,
	typ fuseutil.DirentType,
	ok bool) {
	index, ok := in.findChild(name)
	if ok {
		id = in.entries[index].Inode
		typ = in.entries[index].Type
	}

	return
}

func (in *iNode) findChild(name string) (i int, ok bool) {
	if !in.isDir() {
		panic("findChild called on non-directory.") // TODO
	}

	var e fuseutil.Dirent
	for i, e = range in.entries {
		if e.Name == name {
			ok = true
			return
		}
	}

	return
}

func (in *iNode) readDir(p []byte, offset int) (n int) {
	if !in.isDir() {
		panic("readDir called on non-directory.")
	}

	for i := offset; i < len(in.entries); i++ {
		e := in.entries[i]

		// Skip unused entries.
		if e.Type == fuseutil.DT_Unknown {
			continue
		}

		tmp := fuseutil.WriteDirent(p[n:], in.entries[i])
		if tmp == 0 {
			break
		}

		n += tmp
	}

	return
}

func (in *iNode) readAt(p []byte, off int64) (n int, err error) {
	if !in.isFile() {
		panic("readAt called on non-file.")
	}

	// Ensure the offset is in range.
	if off > int64(len(in.contents)) {
		err = io.EOF
		return
	}

	// Read what we can.
	n = copy(p, in.contents[off:])
	if n < len(p) {
		err = io.EOF
	}

	return
}

func (in *iNode) addChild(
	id fuseops.InodeID,
	name string,
	dt fuseutil.DirentType) {
	var index int

	// Update the modification time.
	in.attrs.Mtime = time.Now()

	// No matter where we place the entry, make sure it has the correct Offset
	// field.
	defer func() {
		in.entries[index].Offset = fuseops.DirOffset(index + 1)
	}()

	// Set up the entry.
	e := fuseutil.Dirent{
		Inode: id,
		Name:  name,
		Type:  dt,
	}

	// Look for a gap in which we can insert it.
	for index = range in.entries {
		if in.entries[index].Type == fuseutil.DT_Unknown {
			in.entries[index] = e
			return
		}
	}

	// Append it to the end.
	index = len(in.entries)
	in.entries = append(in.entries, e)
}

func (in *iNode) writeAt(p []byte, off int64) (n int, err error) {
	if !in.isFile() {
		panic("writeAt called on non-file.")
	}

	// Update the modification time.
	in.attrs.Mtime = time.Now()

	// Ensure that the contents slice is long enough.
	newLen := int(off) + len(p)
	if len(in.contents) < newLen {
		padding := make([]byte, newLen-len(in.contents))
		in.contents = append(in.contents, padding...)
		in.attrs.Size = uint64(newLen)
	}

	// Copy in the data.
	n = copy(in.contents[off:], p)

	// Sanity check.
	if n != len(p) {
		panic(fmt.Sprintf("Unexpected short copy: %v", n))
	}

	return
}

func (in *iNode) isDir() bool {
	return in.attrs.Mode&os.ModeDir != 0
}

func (in *iNode) isSymlink() bool {
	return in.attrs.Mode&os.ModeSymlink != 0
}

func (in *iNode) isFile() bool {
	return !(in.isDir() || in.isSymlink())
}

// Update attributes from non-nil parameters.
func (in *iNode) setAttributes(
	size *uint64,
	mode *os.FileMode,
	mtime *time.Time) {
	// TODO
}
