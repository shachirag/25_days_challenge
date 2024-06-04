package graph

import (
	"challenge/database"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB        *database.DB
	S3Client  *manager.Uploader
	SESClient *ses.Client
}
