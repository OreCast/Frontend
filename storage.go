package main

import (
	"context"
	"fmt"
	"log"

	minio "github.com/minio/minio-go/v7"
	credentials "github.com/minio/minio-go/v7/pkg/credentials"
)

func datasets(uri string) []string {
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
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
