package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ResetPassword(ctx context.Context, db *database.DB, input model.ResetPasswordRequestInput) *model.ResponseModel {
	var (
		userColl = db.GetCollection("user")
		admin     entity.CustomerEntity
	)

	smallEmail := strings.ToLower(input.Email)

	err := userColl.FindOne(ctx, bson.M{"email": smallEmail}).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.ResponseModel{
				Status:  false,
				Message: "No user found",
			}
		}

		return &model.ResponseModel{
			Status:  false,
			Message: "Internal server error, while getting the user: " + err.Error(),
		}
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Failed to hash the password: " + err.Error(),
		}
	}

	_, err = userColl.UpdateOne(ctx, bson.M{"_id": admin.Id}, bson.M{"$set": bson.M{"password": hashedPassword}})
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Failed to update password in the database: " + err.Error(),
		}
	}

	return &model.ResponseModel{
		Status:  true,
		Message: "Password updated successfully after OTP verification",
	}
}
