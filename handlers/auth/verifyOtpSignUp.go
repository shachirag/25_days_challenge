package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"os"
	"time"

	jtoken "github.com/golang-jwt/jwt/v4"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func VerifyOtp(ctx context.Context, db *database.DB, data model.VerifyOtpRequestInput) (*model.LoginResponse, error) {
	otpData, err := fetchLatestOtp(ctx, db, data.Email)
	if err != nil {
		return nil, err
	}

	if data.Otp != otpData.Otp {
		return nil, gqlerror.Errorf("Invalid OTP.")
	}

	userData, err := findOrCreateUser(ctx, db, data)
	if err != nil {
		return nil, err
	}

	if userData == nil {
		return nil, gqlerror.Errorf("Error Occured")
	}

	token, err := GenerateJWTToken(*userData)
	if err != nil {
		return nil, gqlerror.Errorf("Error generating JWT token.")
	}

	taskId, err := createTask(ctx, db, userData.Id)
	if err != nil {
		return nil, err
	}

	challenges := []*model.Challenge{
		{
			ID:             taskId.Hex(),
			UserID:         userData.Id.Hex(),
			Level:          1,
			Day:            1,
			Date:           time.Now().UTC().Format(time.DateOnly),
			CompletedTasks: []string{},
			Status:         "incomplete",
		},
	}

	return &model.LoginResponse{
		User: &model.User{
			ID:       userData.Id.Hex(),
			UserName: userData.UserName,
			Email:    userData.Email,
			Token:    token,
		},
		ChallengeStartDate: time.Now().UTC().Format(time.DateOnly),
		Challenges:         challenges,
	}, nil
}

func fetchLatestOtp(ctx context.Context, db *database.DB, email string) (*entity.OtpEntity, error) {
	otpColl := db.GetCollection("otp")
	var otpData entity.OtpEntity
	err := otpColl.FindOne(ctx, bson.M{"email": email}, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("OTP not found.")
		}
		return nil, gqlerror.Errorf("Error fetching OTP data.")
	}
	return &otpData, nil
}

func findOrCreateUser(ctx context.Context, db *database.DB, userReq model.VerifyOtpRequestInput) (*entity.CustomerEntity, error) {
	customerColl := db.GetCollection("user")
	var userData entity.CustomerEntity
	err := customerColl.FindOne(ctx, bson.M{"email": userReq.Email}).Decode(&userData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, gqlerror.Errorf("Error hashing password.")
			}

			sessionID := primitive.NewObjectID().Hex()
			userId := primitive.NewObjectID()
			userData = entity.CustomerEntity{
				Id:        userId,
				UserName:  userReq.Username,
				Email:     userReq.Email,
				Password:  string(hashedPassword),
				SessionId: sessionID,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			_, err = customerColl.InsertOne(ctx, userData)
			if err != nil {
				return nil, gqlerror.Errorf("Error inserting user data.")
			}
		} else {
			return nil, gqlerror.Errorf("Error fetching user data.")
		}
	}
	return &userData, nil
}

func createTask(ctx context.Context, db *database.DB, userId primitive.ObjectID) (*primitive.ObjectID, error) {

	taskId := primitive.NewObjectID()
	taskData := entity.TasksEntity{
		Id:                taskId,
		UserId:            userId,
		Level:             1,
		Day:               1,
		Status:            "incomplete",
		ChalengeStartDate: time.Now().UTC(),
		Date:              time.Now().UTC(),
		CompletedTasks:    []string{},
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	_, err := db.GetCollection("task").InsertOne(ctx, taskData)
	if err != nil {
		return nil, gqlerror.Errorf("Error inserting task data.")
	}

	return &taskId, nil
}

func GenerateJWTToken(user entity.CustomerEntity) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return "", gqlerror.Errorf("JWT secret key not found.")
	}

	claims := jtoken.MapClaims{
		"Id":        user.Id,
		"email":     user.Email,
		"role":      "customer",
		"sessionId": user.SessionId,
		"exp":       time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
