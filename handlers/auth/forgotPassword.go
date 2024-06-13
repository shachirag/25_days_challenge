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
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ForgotPassword(ctx context.Context, db *database.DB, sesClient *ses.Client, input model.ForgotPasswordRequestInput) (*model.User, error) {

	var (
		userColl = db.GetCollection("user")
		otpColl  = db.GetCollection("otp")
		user     entity.CustomerEntity
	)

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	err := userColl.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusBadRequest, "user not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Internal server error while fetching the user.")
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
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to store OTP in the database: "+err.Error())
	}

	_, err = utils.SendEmail(sesClient, user.UserName, otp)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Internal server error while sending the email: "+err.Error())
	}

	return &model.User{
		ID:       user.Id.Hex(),
		UserName: user.UserName,
		Email:    user.Email,
	}, nil
}
