package azblob

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/aos-dev/go-storage/v3/pkg/iowrap"
	. "github.com/aos-dev/go-storage/v3/types"
)

func (s *Storage) create(path string, opt pairStorageCreate) (o *Object) {
	o = s.newObject(false)
	o.Mode = ModeRead
	o.ID = s.getAbsPath(path)
	o.Path = path
	return o
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

	output, err := s.bucket.NewBlockBlobURL(rp).Download(
		ctx, offset, count,
		azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
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

	output, err := s.bucket.NewBlockBlobURL(rp).GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
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

	sm := make(map[string]string)
	if v := output.AccessTier(); v != "" {
		sm[MetadataAccessTier] = v
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

	_, err = s.bucket.NewBlockBlobURL(rp).Upload(
		ctx, iowrap.SizedReadSeekCloser(r, size),
		headers, azblob.Metadata{}, azblob.BlobAccessConditions{},
		accessTier, azblob.BlobTagsMap{}, azblob.ClientProvidedKeyOptions{},
	)
	if err != nil {
		return 0, err
	}
	return size, nil
}
