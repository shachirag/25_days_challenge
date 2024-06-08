package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ForgotPassword(ctx context.Context, db *database.DB, sesClient *ses.Client, input model.ForgotPasswordRequestInput) *model.ResponseModel {
	var (
		userColl = db.GetCollection("user")
		otpColl   = db.GetCollection("otp")
		user      entity.CustomerEntity
	)

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	err := userColl.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.ResponseModel{
				Status:  false,
				Message: "No User Found",
			}
		}

		return &model.ResponseModel{
			Status:  false,
			Message: "Internal server error, while getting the user: " + err.Error(),
		}
	}

	otp := utils.Generate6DigitOtp()
	otpData := entity.OtpEntity{
		Id:        primitive.NewObjectID(),
		Otp:       otp,
		Email:     smallEmail,
		CreatedAt: time.Now().UTC(),
	}

	_, err = otpColl.InsertOne(ctx, otpData)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Failed to store OTP in the database: " + err.Error(),
		}
	}

	_, err = utils.SendEmail(sesClient, user.UserName, otp)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Internal server error, while sending email: " + err.Error(),
		}
	}

	return &model.ResponseModel{
		Status:  true,
		Message: "Successfully send 6 digit OTP.",
	}
}
