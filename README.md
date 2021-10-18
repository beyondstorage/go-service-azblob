# go-services-azblob

[Azure Blob Storage](https://docs.microsoft.com/en-us/azure/storage/blobs/) service support for [go-storage](https://github.com/beyondstorage/go-storage).


## Notes

**This package has been moved to [go-storage](https://github.com/beyondstorage/go-storage/tree/master/services/azblob).**

```shell
go get go.beyondstorage.io/services/azblob/v3
```

## Install

```go
go get github.com/beyondstorage/go-service-azblob/v2
```

## Usage

```go
import (
	"log"

	_ "github.com/beyondstorage/go-service-azblob/v2"
	"github.com/beyondstorage/go-storage/v4/services"
)

func main() {
	store, err := services.NewStoragerFromString("azblob://container_name/path/to/workdir?credential=hmac:<account_name>:<account_key>&endpoint=https:<account_name>.<endpoint_suffix>")
	if err != nil {
		log.Fatal(err)
	}
	
	// Write data from io.Reader into hello.txt
	n, err := store.Write("hello.txt", r, length)
}
```

- See more examples in [go-storage-example](https://github.com/beyondstorage/go-storage-example).
- Read [more docs](https://beyondstorage.io/docs/go-storage/services/azblob) about go-service-azblob.
