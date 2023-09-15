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

func datasets(s3 S3, bucket string) []string {
	var out []string
	ctx := context.Background()
	// Initialize minio client object.
	minioClient, err := minio.New(s3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.AccessKey, s3.AccessSecret, ""),
		Secure: s3.UseSSL,
	})
	if err != nil {
		log.Println("ERROR", err)
		return out
	}

	//     log.Printf("%#v\n", minioClient) // minioClient is now set up
	if bucket == "" {
		buckets, err := minioClient.ListBuckets(ctx)
		if err != nil {
			log.Println("ERROR", err)
			return out
		}
		for _, bucket := range buckets {
			// fmt.Println(bucket)
			out = append(out, fmt.Sprintf("%s", bucket))
		}
		return out
	}

	// list individual buckets
	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			log.Println("ERROR", object.Err)
			return out
		}
		obj := fmt.Sprintf("%v %s %10d %s\n", object.LastModified, object.ETag, object.Size, object.Key)
		out = append(out, obj)
	}
	return out
}
