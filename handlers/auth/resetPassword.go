package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ResetPassword(ctx context.Context, db *database.DB, input model.ResetPasswordRequestInput) (*model.User, error) {
	var (
		userColl = db.GetCollection("user")
		user    entity.CustomerEntity
	)

	smallEmail := strings.ToLower(input.Email)

	err := userColl.FindOne(ctx, bson.M{"email": smallEmail}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusNotFound, "No user found with the provided email.")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Internal server error while fetching the user: "+err.Error())
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to hash the password: "+err.Error())
	}

	_, err = userColl.UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": bson.M{"password": hashedPassword}})
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update the password in the database: "+err.Error())
	}

	return &model.User{
		ID:       user.Id.Hex(),
		UserName: user.UserName,
		Email:    user.Email,
	}, nil
}
