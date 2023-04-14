/*
Copyright © 2023 Bill Beesley <bill@beesley.dev>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/storage"

	"github.com/bbeesley/fn-push/pkg/zip"
	"github.com/spf13/cobra"
)

// gcpCmd represents the gcp command
var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Upload function assets to Cloud Storage",
	Long: `Zips up function assets and uploads them to Google
	Cloud Storage for use in Cloud Functions.`,
	Run: func(cmd *cobra.Command, args []string) {
		functionData := zip.Create(inputPath, include, exclude, rootDir, symlinkNodeModules)
		ctx := context.Background()

		// Sets your Google Cloud Platform project ID.
		var functionKeyName string
		if versionSuffix != "" {
			functionKeyName = fmt.Sprintf("%s-%s.zip", functionKey, versionSuffix)
		} else {
			functionKeyName = fmt.Sprintf("%s.zip", functionKey)
		}

		// Creates a client.
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()
		for _, bucketName := range buckets {
			object := client.Bucket(bucketName).Object(functionKeyName)
			wc := object.NewWriter(ctx)
			_, err = io.Copy(wc, bytes.NewReader(functionData.Bytes()))
			if err != nil {
				log.Fatalf("Failed to upload file: %v", err)
			}
			err := wc.Close()
			if err != nil {
				log.Fatalf("Writer.Close: %v", err)
			}
			fmt.Println("File uploaded successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(gcpCmd)

	gcpCmd.Flags().StringVarP(&inputPath, "inputPath", "p", ".", "The path to the lambda code and node_modules")
	gcpCmd.Flags().StringArrayVarP(&include, "include", "i", []string{"**"}, "An array of globs defining what to bundle")
	gcpCmd.Flags().StringArrayVarP(&exclude, "exclude", "e", []string{}, "An array of globs defining what not to bundle")
	gcpCmd.Flags().StringVar(&rootDir, "rootDir", "", "An optional path within the zip to save the files to")
	gcpCmd.Flags().StringArrayVarP(&buckets, "buckets", "b", []string{}, "A list of buckets to upload to (same order as the regions please")
	gcpCmd.Flags().StringVarP(&functionKey, "functionKey", "f", "", "The path/filename of the zip file in the bucket (you don't need to add the .zip extension, but remember to include a version string of some sort)")
	gcpCmd.Flags().StringVarP(&versionSuffix, "versionSuffix", "v", "", "An optional string to append to layer and function keys to use as a version indicator")

	err := gcpCmd.MarkFlagRequired("buckets")
	if err != nil {
		log.Fatal("Failed to set buckets flag as required", err)
	}
	err = gcpCmd.MarkFlagRequired("functionKey")
	if err != nil {
		log.Fatal("Failed to set functionKey flag as required", err)
	}
}
