package tests

import (
	"os"
	"testing"

	azblob "github.com/beyondstorage/go-service-azblob/v2"
	ps "github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/google/uuid"
)

func setupTest(t *testing.T) types.Storager {
	t.Log("Setup test for azblob")

	store, err := azblob.NewStorager(
		ps.WithCredential(os.Getenv("STORAGE_AZBLOB_CREDENTIAL")),
		ps.WithName(os.Getenv("STORAGE_AZBLOB_NAME")),
		ps.WithEndpoint(os.Getenv("STORAGE_AZBLOB_ENDPOINT")),
		ps.WithWorkDir("/"+uuid.New().String()+"/"),
	)
	if err != nil {
		t.Errorf("new storager: %v", err)
	}
	return store
}
