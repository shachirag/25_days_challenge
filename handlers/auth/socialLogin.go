package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SocialLoginCustomer(ctx context.Context, db *database.DB, input model.SocialLoginRequestInput) *model.LoginPayload {
	var (
		userColl = db.GetCollection("user")
		customer *entity.CustomerEntity
	)

	err := utils.ValidateSocialId(&input)
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: err.Error(),
		}
	}

	filter := bson.M{}
	if input.Type == "apple" {
		filter = bson.M{"appleId": input.SocialID}
	} else if input.Type == "google" {
		filter = bson.M{"googleId": input.SocialID}
	} else {
		return &model.LoginPayload{
			Status:  false,
			Message: "Invalid social login type",
		}
	}

	var smallEmail string
	if input.Email != nil {
		smallEmail = strings.ToLower(*input.Email)
	}

	err = userColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if input.Email != nil {
			filter = bson.M{"email": smallEmail}
			err = userColl.FindOne(ctx, filter).Decode(&customer)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					customer, err = socialSignup(ctx, db, &input)
					if err != nil {
						return &model.LoginPayload{
							Status:  false,
							Message: "Failed to perform social signup: " + err.Error(),
						}
					}
				} else {
					return &model.LoginPayload{
						Status:  false,
						Message: "Internal server error while getting the user: " + err.Error(),
					}
				}
			}

		} else {
			return &model.LoginPayload{
				Status:  false,
				Message: "Customer not found",
			}
		}
	}

	if customer != nil {
		update := bson.M{}
		isUpdate := false
		if customer.Email == "" && smallEmail != "" {
			update["email"] = input.Email
			isUpdate = true
		}
		if customer.SocialDetails.AppleId == "" && input.Type == "apple" {
			update["socialDetails.appleId"] = input.SocialID
			isUpdate = true
		}
		if customer.SocialDetails.GoogleId == "" && input.Type == "google" {
			update["socialDetails.googleId"] = input.SocialID
			isUpdate = true
		}

		if isUpdate {
			_, err = userColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, bson.M{"$set": update})
			if err != nil {
				return &model.LoginPayload{
					Status:  false,
					Message: "Error updating data: " + err.Error(),
				}
			}
		}
	}

	token, err := generateJWTToken(customer, smallEmail)
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "Failed to generate JWT token: " + err.Error(),
		}
	}

	return &model.LoginPayload{
		Status:  true,
		Message: "Successfully logged in.",
		Data: &model.LoginResponse{
			User: &model.User{
				ID:       customer.Id.Hex(),
				UserName: customer.UserName,
				Email:    customer.Email,
				Token:    token,
			},
		},
	}
}

func socialSignup(ctx context.Context, db *database.DB, data *model.SocialLoginRequestInput) (*entity.CustomerEntity, error) {
	userColl := db.GetCollection("user")

	var smallEmail string
	if data.Email != nil {
		smallEmail = strings.ToLower(*data.Email)
	}

	filter := bson.M{
		"email": smallEmail,
	}

	exists, err := userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	if exists > 0 {
		return nil, fiber.ErrBadRequest
	}

	id := primitive.NewObjectID()

	customer := &entity.CustomerEntity{
		Id:        id,
		Email:     smallEmail,
		UserName:  data.Name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if data.Type == "google" {
		customer.SocialDetails.GoogleId = data.SocialID
	} else if data.Type == "apple" {
		customer.SocialDetails.AppleId = data.SocialID
	} else {
		return nil, fiber.ErrBadRequest
	}

	_, err = userColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func generateJWTToken(customer *entity.CustomerEntity, email string) (string, error) {
	_secret := os.Getenv("JWT_SECRET_KEY")

	month := (time.Hour * 24) * 30
	claims := jwt.MapClaims{
		"Id":    customer.Id,
		"email": email,
		"role":  "customer",
		"exp":   time.Now().Add(month * 6).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(_secret))
}
