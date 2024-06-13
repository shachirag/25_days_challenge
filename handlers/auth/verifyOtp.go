package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func VerifyOtpForResetPassword(ctx context.Context, db *database.DB, input model.VerifyOtpForResetPasswordRequestInput) (*model.User, error) {
	var (
		otpColl = db.GetCollection("otp")
		otpData entity.OtpEntity
		user    entity.CustomerEntity
	)

	if input.Otp == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "OTP is required")
	}

	smallEmail := strings.ToLower(input.Email)

	err := db.GetCollection("user").FindOne(ctx, bson.M{"email": smallEmail}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusBadRequest, "user not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Internal server error while fetching the user.")
	}

	err = otpColl.FindOne(ctx, bson.M{"email": smallEmail}, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusNotFound, "Invalid OTP")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Internal server error while fetching OTP: "+err.Error())
	}

	if input.Otp != otpData.Otp {
		return nil, fiber.NewError(fiber.StatusBadRequest, "OTP does not match")
	}

	return &model.User{
		ID:       user.Id.Hex(),
		UserName: user.UserName,
		Email:    user.Email,
	}, nil
}
