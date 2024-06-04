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
)

func SignUpUser(ctx context.Context, db *database.DB, userInfo model.SignUpRequestInput, sesClient *ses.Client) *model.ResponseModel {
	customerColl := db.GetCollection("user")
	otpColl := db.GetCollection("otp")

	filter := bson.M{
		"email": strings.ToLower(userInfo.Email),
	}

	exists, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Email is already in use.",
		}
	}

	if exists > 0 {
		return &model.ResponseModel{
			Status:  false,
			Message: "Email is already in use.",
		}
	}

	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	// if err != nil {
	// 	return &model.ResponseModel{
	// 		Status:  false,
	// 		Message: "Error hash password",
	// 	}
	// }

	id := primitive.NewObjectID()
	otp := utils.Generate6DigitOtp()
	otpData := entity.OtpEntity{
		Id:        id,
		Otp:       otp,
		Email:     userInfo.Email,
		CreatedAt: time.Now().UTC(),
	}

	_, err = otpColl.InsertOne(ctx, otpData)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Error hash password",
		}
	}

	_, err = utils.SendEmail(sesClient, userInfo.Email, otp)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Error Sending OTP to your email address.",
		}
	}

	return &model.ResponseModel{
		Status:  false,
		Message: "Otp sent successfully.",
	}
}
