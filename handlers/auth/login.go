package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"strings"
	"time"

	jtoken "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func LoginCustomer(ctx context.Context, db *database.DB, input model.LoginRequestInput) *model.LoginPayload {
	customerColl := db.GetCollection("user")

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	var customer entity.CustomerEntity
	err := customerColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.LoginPayload{
				Status:  false,
				Message: "User not found",
			}
		}
		return &model.LoginPayload{
			Status:  false,
			Message: "An error occurred",
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(strings.TrimSpace(input.Password)))
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "Invalid credentials",
		}
	}

	// _secret := os.Getenv("JWT_SECRET_KEY")
	month := (time.Hour * 24) * 30
	claims := jtoken.MapClaims{
		"Id":    customer.Id,
		"email": smallEmail,
		"role":  "customer",
		"exp":   time.Now().Add(month * 6).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	_token, err := token.SignedString([]byte("challenge"))
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "failed to get token",
		}
	}

	return &model.LoginPayload{
		Status:  true,
		Message: "Login successful.",
		LoginResponse: &model.LoginResponse{
			ID:       customer.Id.Hex(),
			Username: customer.UserName,
			Email:    customer.Email,
			Token:    _token,
		},
	}
}
