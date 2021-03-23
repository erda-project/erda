package nexus

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexus_ListBlobStore(t *testing.T) {
	blobStores, err := n.ListBlobStore(BlobStoreListRequest{})
	assert.NoError(t, err)
	printJSON(blobStores)
}

func TestNexus_DeleteBlobStore(t *testing.T) {
	err := n.DeleteBlobStore(BlobStoreDeleteRequest{
		BlobName: "a",
	})
	assert.NoError(t, err)
}

func TestNexus_EnsureFileBlobStore(t *testing.T) {
	err := n.EnsureFileBlobStore(FileBlobStoreCreateRequest{
		SoftQuota: nil,
		Path:      "test-blob-0410",
		Name:      "test-blob-0410",
	})
	assert.NoError(t, err)
}

func TestNexus_CreateFileBlobStore(t *testing.T) {
	err := n.CreateFileBlobStore(FileBlobStoreCreateRequest{
		SoftQuota: nil,
		Path:      "maven-blob-1",
		Name:      "maven-blob-1",
	})
	assert.NoError(t, err)
}

func TestNexus_GetFileBlobStore(t *testing.T) {
	store, err := n.GetFileBlobStore(FileBlobStoreGetRequest{
		Name: "test-blob-100",
	})
	assert.NoError(t, err)
	printJSON(store)
}

func TestNexus_UpdateFileBlobStore(t *testing.T) {
	err := n.UpdateFileBlobStore(FileBlobStoreUpdateRequest{FileBlobStoreCreateRequest{
		Name: "test-blob-1",
		Path: "string",
	}})
	assert.NoError(t, err)
}

func TestFileBlobStoreCreateRequest_HandlePath(t *testing.T) {
	oriPath := "docker-hosted-platform"
	req := FileBlobStoreCreateRequest{
		Path:           oriPath,
		BlobUseNetdata: BlobUseNetdata{UseNetdata: false},
	}
	req = req.handlePath(DefaultBlobNetdataDir)
	assert.Equal(t, oriPath, req.Path)
	fmt.Println(req.Path)

	// use netdata
	req.UseNetdata = true
	req = req.handlePath(DefaultBlobNetdataDir)
	assert.Equal(t, filepath.Join(DefaultBlobNetdataDir, oriPath), req.Path)
	fmt.Println(req.Path)
}
