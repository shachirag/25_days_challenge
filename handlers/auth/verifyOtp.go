package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func VerifyOtpForResetPassword(ctx context.Context, db *database.DB, input model.VerifyOtpForResetPasswordRequestInput) *model.ResponseModel {
	var (
		otpColl = db.GetCollection("otp")
		otpData entity.OtpEntity
	)

	if input.Otp == "" {
		return &model.ResponseModel{
			Status:  false,
			Message: "Entered OTP is required",
		}
	}

	smallEmail := strings.ToLower(input.Email)

	err := otpColl.FindOne(ctx, bson.M{"email": smallEmail}, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.ResponseModel{
				Status:  false,
				Message: "Invalid OTP",
			}
		}

		return &model.ResponseModel{
			Status:  false,
			Message: "Internal server error, while getting the user: " + err.Error(),
		}
	}

	if input.Otp != otpData.Otp {
		return &model.ResponseModel{
			Status:  false,
			Message: "Invalid OTP",
		}
	}

	return &model.ResponseModel{
		Status:  true,
		Message: "OTP verified successfully",
	}
}
