package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"os"
	"time"

	jtoken "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func VerifyOtp(ctx context.Context, db *database.DB, otpInfo model.VerifyOtpRequestInput) *model.LoginPayload {
	var (
		customerColl = db.GetCollection("user")
		otpColl      = db.GetCollection("otp")
		otpData      entity.OtpEntity
	)

	err := otpColl.FindOne(ctx, bson.M{"email": otpInfo.Email}, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.LoginPayload{
				Status:  false,
				Message: "Failed to find OTP.",
			}
		}

		return &model.LoginPayload{
			Status:  false,
			Message: "Internal server error while fetching user.",
		}
	}
	if otpInfo.Otp != otpData.Otp {
		return &model.LoginPayload{
			Status:  false,
			Message: "Invalid OTP.",
		}
	}

	var userData entity.CustomerEntity

	err = customerColl.FindOne(ctx, bson.M{"email": otpInfo.Email}).Decode(&userData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			userData = entity.CustomerEntity{
				Email: otpInfo.Email,
			}
		} else {
			return &model.LoginPayload{
				Status:  false,
				Message: "Internal server error while fetching user.",
			}
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(otpInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "Error hash password",
		}
	}

	id := primitive.NewObjectID()
	userData = entity.CustomerEntity{
		Id:       id,
		UserName: otpInfo.Username,
		Email:    otpInfo.Email,
		Password: string(hashedPassword),
		// UserName:  otpInfo.Username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err = customerColl.InsertOne(ctx, userData)
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "Failed to Insert data.",
		}
	}

	_secret := os.Getenv("JWT_SECRET_KEY")
	month := (time.Hour * 24) * 30
	claims := jtoken.MapClaims{
		"Id":    userData.Id,
		"email": userData.Email,
		"exp":   time.Now().Add(month * 6).Unix(),
		"role":  "customer",
	}
	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	_token, err := token.SignedString([]byte(_secret))
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "Token is Invalid.",
		}
	}

	return &model.LoginPayload{
		Status:  true,
		Message: "Otp verified successful.",
		LoginResponse: &model.LoginResponse{
			ID:       userData.Id.Hex(),
			Username: userData.UserName,
			Email:    userData.Email,
			Token:    _token,
		},
	}
}
