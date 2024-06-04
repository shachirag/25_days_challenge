package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var (
	charSet = aws.String("UTF-8")
	sender  = aws.String("selfchallenge@yopmail.com")
	subject = aws.String("OTP for reset password")
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
