package utils

// import (
// 	"challenge/database"
// 	"context"
// 	"fmt"
// 	"mime/multipart"
// 	"os"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// )

// func UploadToS3(fName string, file multipart.File) (string, error) {
// 	bucketName := os.Getenv("S3_BUCKET_NAME")
// 	if bucketName == "" {
// 		return "", fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
// 	}

// 	uploader := database.GetS3Uploader()
// 	if uploader == nil {
// 		return "", fmt.Errorf("failed to get S3 uploader")
// 	}

// 	u, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(fName),
// 		Body:   file,
// 		ACL:    "public-read",
// 	})
// 	if err != nil {
// 		return "", fmt.Errorf("failed to upload file to S3: %v", err.Error())
// 	}

// 	return u.Location, nil
// }
