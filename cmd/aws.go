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
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bbeesley/fn-push/pkg/zip"
	"github.com/spf13/cobra"
)

func S3Upload(region string, bucket string, keyName string, functionData *bytes.Buffer) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = region
	})

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keyName),
		Body:   bytes.NewReader(functionData.Bytes()),
	})
	if err != nil {
		log.Fatalf("failed to upload file '%s'", keyName)
	}
	fmt.Printf("Successfully uploaded %s to %s in %s\n", keyName, bucket, region)
}

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Upload lambda assets to S3",
	Long: `Zips up function assets and uploads them to AWS S3
	for use in lambda functions. Optionally creates a file for
	a layer as well as a file for the function itself.`,
	Run: func(cmd *cobra.Command, args []string) {
		var functionKeyName string
		if versionSuffix != "" {
			functionKeyName = fmt.Sprintf("%s-%s.zip", functionKey, versionSuffix)
		} else {
			functionKeyName = fmt.Sprintf("%s.zip", functionKey)
		}

		if layerKey == "" {
			functionData := zip.Create(inputPath, include, exclude, rootDir, symlinkNodeModules)
			for ix, region := range regions {
				S3Upload(region, buckets[ix], functionKeyName, functionData)
			}
		} else {
			functionExclude := exclude
			if symlinkNodeModules {
				functionExclude = append(functionExclude, "node_modules/**")
			}
			functionData := zip.Create(inputPath, include, functionExclude, rootDir, symlinkNodeModules)
			layerData := zip.Create(inputPath, []string{"node_modules/**"}, []string{}, rootDir, false)
			var layerKeyName string
			if versionSuffix != "" {
				layerKeyName = fmt.Sprintf("%s-%s.zip", layerKey, versionSuffix)
			} else {
				layerKeyName = fmt.Sprintf("%s.zip", layerKey)
			}
			for ix, region := range regions {
				S3Upload(region, buckets[ix], functionKeyName, functionData)
				S3Upload(region, buckets[ix], layerKeyName, layerData)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)
	awsCmd.Flags().StringVarP(&inputPath, "inputPath", "p", ".", "The path to the lambda code and node_modules")
	awsCmd.Flags().StringArrayVarP(&include, "include", "i", []string{"**"}, "An array of globs defining what to bundle")
	awsCmd.Flags().StringArrayVarP(&exclude, "exclude", "e", []string{}, "An array of globs defining what not to bundle")
	awsCmd.Flags().StringVar(&rootDir, "rootDir", "", "An optional path within the zip to save the files to")
	awsCmd.Flags().StringArrayVarP(&regions, "regions", "r", []string{}, "A list of regions to upload the assets in")
	awsCmd.Flags().StringArrayVarP(&buckets, "buckets", "b", []string{}, "A list of buckets to upload to (same order as the regions please")
	awsCmd.Flags().StringVarP(&functionKey, "functionKey", "f", "", "The path/filename of the zip file in the bucket (you don't need to add the .zip extension, but remember to include a version string of some sort)")
	awsCmd.Flags().StringVarP(&layerKey, "layerKey", "l", "", "Tells the module to split out the node modules into a zip that you can create a lambda layer from")
	awsCmd.Flags().StringVarP(&versionSuffix, "versionSuffix", "v", "", "An optional string to append to layer and function keys to use as a version indicator")
	awsCmd.Flags().BoolVarP(&symlinkNodeModules, "symlinkNodeModules", "n", false, "Should we create a symlink from the function directory to the layer node_modules?")

	err := awsCmd.MarkFlagRequired("regions")
	if err != nil {
		log.Fatal("Failed to set regions flag as required", err)
	}
	err = awsCmd.MarkFlagRequired("buckets")
	if err != nil {
		log.Fatal("Failed to set buckets flag as required", err)
	}
	err = awsCmd.MarkFlagRequired("functionKey")
	if err != nil {
		log.Fatal("Failed to set functionKey flag as required", err)
	}
}
