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

// SocialLoginCustomer handles social login or signup for a customer
func SocialLoginCustomer(ctx context.Context, db *database.DB, input model.SocialLoginRequestInput) (*model.LoginResponse, error) {
	var (
		userColl = db.GetCollection("user")
		customer *entity.CustomerEntity
	)

	// Validate the social ID input
	err := utils.ValidateSocialId(&input)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid social ID: "+err.Error())
	}

	// Prepare the filter based on social login type
	filter := bson.M{}
	switch input.Type {
	case "apple":
		filter = bson.M{"appleId": input.SocialID}
	case "google":
		filter = bson.M{"googleId": input.SocialID}
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Unsupported social login type")
	}

	// Lowercase email for uniformity
	var smallEmail string
	if input.Email != nil {
		smallEmail = strings.ToLower(*input.Email)
	}

	// Attempt to find the user by social ID
	err = userColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if input.Email != nil {
			// If not found by social ID, try finding by email
			filter = bson.M{"email": smallEmail}
			err = userColl.FindOne(ctx, filter).Decode(&customer)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// No document found, attempt to sign up
					customer, err = socialSignup(ctx, db, &input)
					if err != nil {
						return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to sign up: "+err.Error())
					}
				} else {
					return nil, fiber.NewError(fiber.StatusInternalServerError, "Error finding user by email: "+err.Error())
				}
			}
		} else {
			// Email not provided, user not found
			return nil, fiber.NewError(fiber.StatusNotFound, "User not found with provided social ID and no email to fallback")
		}
	}

	// Update user details if necessary
	if customer != nil {
		update := bson.M{}
		isUpdate := false
		if customer.Email == "" && smallEmail != "" {
			update["email"] = smallEmail
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
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update user details: "+err.Error())
			}
		}
	}

	// Generate JWT token for the authenticated user
	token, err := generateJWTToken(customer, smallEmail)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to generate JWT token: "+err.Error())
	}

	// Return the login payload with user information and token
	return &model.LoginResponse{
		User: &model.User{
			ID:       customer.Id.Hex(),
			UserName: customer.UserName,
			Email:    customer.Email,
			Token:    token,
		},
	}, nil
}

// Helper function to handle social signup
func socialSignup(ctx context.Context, db *database.DB, data *model.SocialLoginRequestInput) (*entity.CustomerEntity, error) {
	userColl := db.GetCollection("user")

	var smallEmail string
	if data.Email != nil {
		smallEmail = strings.ToLower(*data.Email)
	}

	// Check if a user with the same email already exists
	filter := bson.M{"email": smallEmail}
	exists, err := userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error checking existing user: "+err.Error())
	}
	if exists > 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "User with this email already exists")
	}

	// Create new user entity
	id := primitive.NewObjectID()
	customer := &entity.CustomerEntity{
		Id:        id,
		Email:     smallEmail,
		UserName:  data.Name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Set the social details based on the login type
	switch data.Type {
	case "google":
		customer.SocialDetails.GoogleId = data.SocialID
	case "apple":
		customer.SocialDetails.AppleId = data.SocialID
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Unsupported social login type")
	}

	// Insert the new customer into the database
	_, err = userColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to insert new customer: "+err.Error())
	}

	return customer, nil
}

// Helper function to generate a JWT token
func generateJWTToken(customer *entity.CustomerEntity, email string) (string, error) {
	_secret := os.Getenv("JWT_SECRET_KEY")

	month := (time.Hour * 24) * 30
	claims := jwt.MapClaims{
		"Id":    customer.Id.Hex(),
		"email": email,
		"role":  "customer",
		"exp":   time.Now().Add(month * 6).Unix(),
	}

	// Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(_secret))
}
