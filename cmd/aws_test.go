package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var s3BucketName string = "fn-push-testing"
var region string = "eu-west-2"

func TestAWSUpload(t *testing.T) {
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
	S3Upload(region, bucketName, key, &b)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = region
	})
	file, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		panic(err)
	}
	result, err := io.ReadAll(file.Body)
	if err != nil {
		panic(err)
	}
	if string(result) != fileContentText.String() {
		t.Fatalf("Expected: %s, actual: %s", fileContentText.String(), result)
	}
}
