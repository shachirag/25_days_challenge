package utils

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var (
	senderEmail = os.Getenv("SENDER_EMAIL")
	charSet     = aws.String("UTF-8")
	sender      = aws.String(senderEmail)
	subject     = aws.String("OTP for reset password")
)

func SendEmail(sesClient *ses.Client, to string, link string) (*ses.SendEmailOutput, error) {

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				"selfchallenge@yopmail.com",
			},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data:    OtpEmailBodyHtml(link),
					Charset: charSet,
				},
				Text: &types.Content{
					Data:    OtpEmailBodyText(link),
					Charset: charSet,
				},
			},
			Subject: &types.Content{
				Data:    subject,
				Charset: charSet,
			},
		},
		Source: sender,
	}

	return sesClient.SendEmail(context.Background(), input)
}
