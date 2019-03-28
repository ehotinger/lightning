# lightningfs

[![Build Status](https://travis-ci.com/ehotinger/lightningfs.svg?branch=master)](https://travis-ci.com/ehotinger/lightningfs)

## About

Don't use this repository yet. This is mostly a playground for experimentation. Mind the dust and watch out for falling objects.

## Dependencies

### Bazil

- API Docs: https://godoc.org/bazil.org/fuse
- https://bazil.org/fuse/index.html
- https://github.com/bazil/fuse

### azblob

- API Docs: https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob


### Design

- Consider implementing `Getattr` - if the method isn't implemented, the attributes are generated based on `Attr()` with zero values filled in. Perhaps we could do better caching with a custom implementation.