package main

import (
	"context"
	"fmt"
	"log"

	minio "github.com/minio/minio-go/v7"
	credentials "github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 represent S3 storage record
type S3 struct {
	Endpoint     string
	AccessKey    string
	AccessSecret string
	UseSSL       bool
}

func datasets(s3 S3) []string {
	// Initialize minio client object.
	minioClient, err := minio.New(s3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.AccessKey, s3.AccessSecret, ""),
		Secure: s3.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	var out []string

	//     log.Printf("%#v\n", minioClient) // minioClient is now set up
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		fmt.Println(err)
		return out
	}
	for _, bucket := range buckets {
		// fmt.Println(bucket)
		out = append(out, fmt.Sprintf("%s", bucket))
	}
	return out
}
