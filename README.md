[![Build Status](https://github.com/beyondstorage/go-service-azblob/workflows/Unit%20Test/badge.svg?branch=master)](https://github.com/beyondstorage/go-service-azblob/actions?query=workflow%3A%22Unit+Test%22)
[![Integration Tests](https://teamcity.beyondstorage.io/app/rest/builds/buildType:(id:Services_GoServiceAzblob_IntegrationTests)/statusIcon)](https://teamcity.beyondstorage.io/buildConfiguration/Services_GoServiceAzblob_IntegrationTests)
[![License](https://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/Xuanwo/storage/blob/master/LICENSE)
[![](https://img.shields.io/matrix/beyondstorage@go-storage:matrix.org.svg?logo=matrix)](https://matrix.to/#/#beyondstorage@go-storage:matrix.org)

# go-services-azblob

[Azure Blob Storage](https://docs.microsoft.com/en-us/azure/storage/blobs/) service support for [go-storage](https://github.com/beyondstorage/go-storage).

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
