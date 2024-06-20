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
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SignUpUser(ctx context.Context, db *database.DB, sesClient *ses.Client, userInfo model.SignUpRequestInput) (*model.User, error) {
	customerColl := db.GetCollection("user")
	otpColl := db.GetCollection("otp")

	filter := bson.M{
		"email": strings.ToLower(userInfo.Email),
	}

	exists, err := customerColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Database error: " + err.Error())
	}

	if exists > 0 {
		return nil, gqlerror.Errorf("Email is already in use.")
	}

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
		return nil, gqlerror.Errorf("Error storing OTP: " + err.Error())
	}

	_, err = utils.SendEmail(sesClient, userInfo.Email, otp)
	if err != nil {
		return nil, gqlerror.Errorf("Error sending OTP to email: " + err.Error())
	}

	return &model.User{}, nil
}
