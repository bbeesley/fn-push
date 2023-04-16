package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

var bucketName string = "fn-push-testing"

func TestGCPUpload(t *testing.T) {
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("failed to create uuid", err)
	}
	fileContentText, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("failed to create uuid", err)
	}
	key := fmt.Sprintf("%s.txt", id)
	var b bytes.Buffer
	b.WriteString(fileContentText.String())
	StorageUpload(bucketName, key, &b)

	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	r, err := bucket.Object(key).NewReader(ctx)
	if err != nil {
		t.Fatalf("Failed to open object: %v", err)
	}
	defer r.Close()
	result, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read object: %v", err)
	}
	if string(result) != fileContentText.String() {
		t.Fatalf("Expected: %s, actual: %s", fileContentText.String(), result)
	}
}
