package fs

import (
	fuseFS "bazil.org/fuse/fs"
)

type LightningFS struct{}

func NewLightningFS() (*LightningFS, error) {
	return nil, nil
}

// TODO:
// https://godoc.org/bazil.org/fuse/fs#FS
func (fs *LightningFS) Root() (fuseFS.Node, error) {
	return nil, nil
}
