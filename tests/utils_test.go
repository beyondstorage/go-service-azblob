package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/beyondstorage/go-service-azblob/v2"
	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
)

func setupTest(t *testing.T) types.Storager {
	t.Log("Setup test for azblob")

	var connStr strings.Builder
	connStr.WriteString("azblob://")
	connStr.WriteString(os.Getenv("STORAGE_AZBLOB_NAME"))
	connStr.WriteString("/" + uuid.New().String())
	connStr.WriteString("?credential=" + os.Getenv("STORAGE_AZBLOB_CREDENTIAL"))
	connStr.WriteString("&endpoint=" + os.Getenv("STORAGE_AZBLOB_ENDPOINT"))

	store, err := services.NewStoragerFromString(connStr.String())
	if err != nil {
		t.Errorf("new storager: %v", err)
	}
	return store
}
