package fs

import (
	"os"
	"testing"
	"time"

	"github.com/jacobsa/fuse/fuseops"
)

func TestEquals(t *testing.T) {
	var t1 = time.Now()

	for _, test := range []struct {
		attrs    fuseops.InodeAttributes
		node     *iNode
		expected bool
	}{
		{
			attrs: fuseops.InodeAttributes{
				Mode:   0700 | os.ModeDir,
				Uid:    0,
				Gid:    0,
				Mtime:  t1,
				Crtime: t1,
				Atime:  t1,
			},
			node: &iNode{
				attrs: fuseops.InodeAttributes{
					Mode:   0700 | os.ModeDir,
					Uid:    0,
					Gid:    0,
					Mtime:  t1,
					Crtime: t1,
					Atime:  t1,
				},
				xattrs: make(map[string][]byte),
			},
			expected: true,
		},
	} {
		iNode := newINode(test.attrs)
		if actual := iNode.equals(test.node); actual != test.expected {
			t.Fatalf("expected %v but got %v", test.expected, actual)
		}
	}
}
