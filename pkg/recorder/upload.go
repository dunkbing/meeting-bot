//go:build !test
// +build !test

package recorder

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dunkbing/meeting-bot/pkg/config"
	"io"
	"net/url"
	"os"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// TODO: write to persistent volume, use separate upload process

func (r *Recorder) uploadS3() error {
	conf, _ := config.GetConfig()
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			conf.FileOutput.S3.AccessKey,
			conf.FileOutput.S3.Secret,
			"",
		),
		Endpoint: aws.String(conf.FileOutput.S3.Endpoint),
		Region:   aws.String(conf.FileOutput.S3.Region),
	})
	if err != nil {
		return err
	}

	file, err := os.Open(r.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	if _, err = file.Read(buffer); err != nil {
		return err
	}

	_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(conf.FileOutput.S3.Bucket),
		Key:           aws.String(r.filepath),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("video/mp4"),
	})

	return err
}

func (r *Recorder) uploadAzure() error {
	conf, _ := config.GetConfig()
	credential, err := azblob.NewSharedKeyCredential(
		conf.FileOutput.AzBlob.AccountName,
		conf.FileOutput.AzBlob.AccountKey,
	)
	if err != nil {
		return err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s",
			conf.FileOutput.AzBlob.AccountName,
			conf.FileOutput.AzBlob.ContainerName,
		),
	)

	containerURL := azblob.NewContainerURL(*URL, p)

	blobURL := containerURL.NewBlockBlobURL(r.filepath)
	file, err := os.Open(r.filepath)
	if err != nil {
		return err
	}
	// upload blocks in parallel for optimal performance
	// it calls PutBlock/PutBlockList for files larger than 256 MBs and PutBlob for smaller files
	ctx := context.Background()
	_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{ContentType: "video/mp4"},
		BlockSize:       4 * 1024 * 1024,
		Parallelism:     16,
	})
	return err
}

func (r *Recorder) uploadGCP() error {
	conf, _ := config.GetConfig()
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	file, err := os.Open(r.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	wc := client.Bucket(conf.FileOutput.GCPConfig.Bucket).Object(r.filepath).NewWriter(ctx)

	if _, err = io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}
