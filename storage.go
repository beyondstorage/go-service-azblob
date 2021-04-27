package azblob

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/aos-dev/go-storage/v3/pkg/iowrap"
	. "github.com/aos-dev/go-storage/v3/types"
)

func (s *Storage) commitAppend(ctx context.Context, o *Object, opt pairStorageCommitAppend) (err error) {
	return
}

func (s *Storage) create(path string, opt pairStorageCreate) (o *Object) {
	o = s.newObject(false)
	o.Mode = ModeRead
	o.ID = s.getAbsPath(path)
	o.Path = path
	return o
}

func (s *Storage) createAppend(ctx context.Context, path string, opt pairStorageCreateAppend) (o *Object, err error) {
	rp := s.getAbsPath(path)

	var cpk azblob.ClientProvidedKeyOptions
	if opt.HasEncryptionKey {
		cpk, err = calculateEncryptionHeaders(opt.EncryptionKey, opt.EncryptionScope)
		if err != nil {
			return
		}
	}
	_, err = s.bucket.NewAppendBlobURL(rp).Create(ctx, azblob.BlobHTTPHeaders{}, nil,
		azblob.BlobAccessConditions{}, nil, cpk)
	if err != nil {
		return
	}

	o = s.newObject(true)
	o.Mode = ModeRead | ModeAppend
	o.ID = rp
	o.Path = path
	o.SetAppendOffset(0)
	return o, nil
}

func (s *Storage) delete(ctx context.Context, path string, opt pairStorageDelete) (err error) {
	rp := s.getAbsPath(path)

	_, err = s.bucket.NewBlockBlobURL(rp).Delete(ctx,
		azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) list(ctx context.Context, path string, opt pairStorageList) (oi *ObjectIterator, err error) {
	input := &objectPageStatus{
		maxResults: 200,
		prefix:     s.getAbsPath(path),
	}

	var nextFn NextObjectFunc

	switch {
	case opt.ListMode.IsDir():
		input.delimiter = "/"
		nextFn = s.nextObjectPageByDir
	case opt.ListMode.IsPrefix():
		nextFn = s.nextObjectPageByPrefix
	default:
		return nil, fmt.Errorf("invalid list mode")
	}

	return NewObjectIterator(ctx, nextFn, input), nil
}

func (s *Storage) metadata(ctx context.Context, opt pairStorageMetadata) (meta *StorageMeta, err error) {
	meta = NewStorageMeta()
	meta.Name = s.name
	meta.WorkDir = s.workDir
	return meta, nil
}

func (s *Storage) nextObjectPageByDir(ctx context.Context, page *ObjectPage) error {
	input := page.Status.(*objectPageStatus)

	output, err := s.bucket.ListBlobsHierarchySegment(ctx, input.marker, input.delimiter, azblob.ListBlobsSegmentOptions{
		Prefix:     input.prefix,
		MaxResults: input.maxResults,
	})
	if err != nil {
		return err
	}

	for _, v := range output.Segment.BlobPrefixes {
		o := s.newObject(true)
		o.ID = v.Name
		o.Path = s.getRelPath(v.Name)
		o.Mode |= ModeDir

		page.Data = append(page.Data, o)
	}

	for _, v := range output.Segment.BlobItems {
		o, err := s.formatFileObject(v)
		if err != nil {
			return err
		}

		page.Data = append(page.Data, o)
	}

	if !output.NextMarker.NotDone() {
		return IterateDone
	}

	input.marker = output.NextMarker
	return nil
}

func (s *Storage) nextObjectPageByPrefix(ctx context.Context, page *ObjectPage) error {
	input := page.Status.(*objectPageStatus)

	output, err := s.bucket.ListBlobsFlatSegment(ctx, input.marker, azblob.ListBlobsSegmentOptions{
		Prefix:     input.prefix,
		MaxResults: input.maxResults,
	})
	if err != nil {
		return err
	}

	for _, v := range output.Segment.BlobItems {
		o, err := s.formatFileObject(v)
		if err != nil {
			return err
		}

		page.Data = append(page.Data, o)
	}

	if !output.NextMarker.NotDone() {
		return IterateDone
	}

	input.marker = output.NextMarker
	return nil
}

func (s *Storage) read(ctx context.Context, path string, w io.Writer, opt pairStorageRead) (n int64, err error) {
	rp := s.getAbsPath(path)

	offset := int64(0)
	if opt.HasOffset {
		offset = opt.Offset
	}

	count := int64(azblob.CountToEnd)
	if opt.HasSize {
		count = opt.Size
	}

	var cpk azblob.ClientProvidedKeyOptions
	if opt.HasEncryptionKey {
		cpk, err = calculateEncryptionHeaders(opt.EncryptionKey, opt.EncryptionScope)
		if err != nil {
			return 0, err
		}
	}
	output, err := s.bucket.NewBlockBlobURL(rp).Download(
		ctx, offset, count,
		azblob.BlobAccessConditions{}, false, cpk)
	if err != nil {
		return 0, err
	}
	defer func() {
		cerr := output.Response().Body.Close()
		if cerr != nil {
			err = cerr
		}
	}()

	rc := output.Response().Body
	if opt.HasIoCallback {
		rc = iowrap.CallbackReadCloser(rc, opt.IoCallback)
	}

	return io.Copy(w, rc)
}

func (s *Storage) stat(ctx context.Context, path string, opt pairStorageStat) (o *Object, err error) {
	rp := s.getAbsPath(path)

	var cpk azblob.ClientProvidedKeyOptions
	if opt.HasEncryptionKey {
		cpk, err = calculateEncryptionHeaders(opt.EncryptionKey, opt.EncryptionScope)
		if err != nil {
			return
		}
	}

	output, err := s.bucket.NewBlockBlobURL(rp).GetProperties(ctx, azblob.BlobAccessConditions{}, cpk)
	if err != nil {
		return nil, err
	}

	o = s.newObject(true)
	o.ID = rp
	o.Path = path
	o.Mode |= ModeRead

	o.SetContentLength(output.ContentLength())
	o.SetLastModified(output.LastModified())

	if v := string(output.ETag()); v != "" {
		o.SetEtag(v)
	}
	if v := output.ContentType(); v != "" {
		o.SetContentType(v)
	}
	if v := output.ContentMD5(); len(v) > 0 {
		o.SetContentMd5(base64.StdEncoding.EncodeToString(v))
	}

	var sm ObjectMetadata
	if v := output.AccessTier(); v != "" {
		sm.AccessTier = v
	}
	if v := output.EncryptionKeySha256(); v != "" {
		sm.EncryptionKeySha256 = v
	}
	if v := output.EncryptionScope(); v != "" {
		sm.EncryptionScope = v
	}
	if v, err := strconv.ParseBool(output.IsServerEncrypted()); err == nil {
		sm.ServerEncrypted = v
	}
	o.SetServiceMetadata(sm)

	return o, nil
}

func (s *Storage) write(ctx context.Context, path string, r io.Reader, size int64, opt pairStorageWrite) (n int64, err error) {
	rp := s.getAbsPath(path)

	if opt.HasIoCallback {
		r = iowrap.CallbackReader(r, opt.IoCallback)
	}

	accessTier := azblob.AccessTierNone
	if opt.HasAccessTier {
		accessTier = azblob.AccessTierType(opt.AccessTier)
	}

	headers := azblob.BlobHTTPHeaders{}
	if opt.HasContentMd5 {
		headers.ContentMD5, err = base64.StdEncoding.DecodeString(opt.ContentMd5)
		if err != nil {
			return 0, err
		}
	}
	if opt.HasContentType {
		headers.ContentType = opt.ContentType
	}

	var cpk azblob.ClientProvidedKeyOptions
	if opt.HasEncryptionKey {
		cpk, err = calculateEncryptionHeaders(opt.EncryptionKey, opt.EncryptionScope)
		if err != nil {
			return 0, err
		}
	}
	_, err = s.bucket.NewBlockBlobURL(rp).Upload(
		ctx, iowrap.SizedReadSeekCloser(r, size),
		headers, azblob.Metadata{}, azblob.BlobAccessConditions{},
		accessTier, azblob.BlobTagsMap{}, cpk)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (s *Storage) writeAppend(ctx context.Context, o *Object, r io.Reader, size int64, opt pairStorageWriteAppend) (n int64, err error) {
	rp := o.GetID()

	offset, ok := o.GetAppendOffset()
	if !ok {
		err = fmt.Errorf("append offset is not set")
		return
	}

	var cpk azblob.ClientProvidedKeyOptions
	if opt.HasEncryptionKey {
		cpk, err = calculateEncryptionHeaders(opt.EncryptionKey, opt.EncryptionScope)
		if err != nil {
			return
		}
	}

	var accessConditions azblob.AppendBlobAccessConditions
	if 0 == offset {
		accessConditions.AppendPositionAccessConditions.IfAppendPositionEqual = -1
	} else {
		accessConditions.AppendPositionAccessConditions.IfAppendPositionEqual = offset
	}

	appendResp, err := s.bucket.NewAppendBlobURL(rp).AppendBlock(
		ctx, iowrap.SizedReadSeekCloser(r, size),
		accessConditions, nil, cpk)
	if err != nil {
		return
	}

	// BlobAppendOffset() returns the offset at which the block was committed, in bytes, but seems not the next append position.
	// ref: https://github.com/Azure/azure-storage-blob-go/blob/master/azblob/zt_url_append_blob_test.go
	offset, err = strconv.ParseInt(appendResp.BlobAppendOffset(), 10, 64)
	if err != nil {
		return
	}
	offset += size
	o.SetAppendOffset(offset)

	return offset, nil
}
