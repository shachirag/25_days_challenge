package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	jtoken "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func VerifyOtp(ctx context.Context, db *database.DB, data model.VerifyOtpRequestInput) (*model.LoginResponse, error) {
	otpData, err := fetchLatestOtp(ctx, db, data.User.Email)
	if err != nil {
		return nil, err
	}

	if data.User.Otp != otpData.Otp {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid OTP.")
	}

	userData, err := findOrCreateUser(ctx, db, data.User)
	if err != nil {
		return nil, err
	}

	if userData == nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error Occured")
	}

	token, err := GenerateJWTToken(*userData)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error generating JWT token.")
	}

	taskId, err := createTask(ctx, db, userData.Id, data.Task, data.ChallengeStartDate)
	if err != nil {
		return nil, err
	}

	completedTasks := []string{}
	if len(data.Task.CompletedTasks) > 0 {
		completedTasks = data.Task.CompletedTasks
	}

	challenges := []*model.Challenge{
		{
			ID:             taskId.Hex(),
			UserID:         userData.Id.Hex(),
			Level:          data.Task.Level,
			Day:            data.Task.Day,
			Date:           data.Task.Date,
			CompletedTasks: completedTasks,
			Status:         data.Task.Status,
		},
	}

	return &model.LoginResponse{
		User: &model.User{
			ID:       userData.Id.Hex(),
			UserName: userData.UserName,
			Email:    userData.Email,
			Token:    token,
		},
		ChallengeStartDate: data.ChallengeStartDate,
		Challenges:         challenges,
	}, nil
}

func fetchLatestOtp(ctx context.Context, db *database.DB, email string) (*entity.OtpEntity, error) {
	otpColl := db.GetCollection("otp")
	var otpData entity.OtpEntity
	err := otpColl.FindOne(ctx, bson.M{"email": email}, options.FindOne().SetSort(bson.M{"createdAt": -1})).Decode(&otpData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusNotFound, "OTP not found.")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error fetching OTP data.")
	}
	return &otpData, nil
}

func findOrCreateUser(ctx context.Context, db *database.DB, userReq *model.UserReq) (*entity.CustomerEntity, error) {
	customerColl := db.GetCollection("user")
	var userData entity.CustomerEntity
	err := customerColl.FindOne(ctx, bson.M{"email": userReq.Email}).Decode(&userData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Error hashing password.")
			}

			userId := primitive.NewObjectID()
			userData = entity.CustomerEntity{
				Id:       userId,
				UserName: userReq.Username,
				Email:    userReq.Email,
				Password: string(hashedPassword),
				SocialDetails: entity.SocialDetails{
					AppleId:  userReq.AppleID,
					GoogleId: userReq.GoogleID,
				},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			_, err = customerColl.InsertOne(ctx, userData)
			if err != nil {
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Error inserting user data.")
			}
		} else {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Error fetching user data.")
		}
	}
	return &userData, nil
}

func createTask(ctx context.Context, db *database.DB, userId primitive.ObjectID, task *model.TaskInput, challengeStartDate string) (*primitive.ObjectID, error) {
	var date time.Time
	if task.Date != "" {
		var err error
		date, err = time.Parse(time.DateOnly, task.Date)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to parse date")
		}
	} else {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Date is mandatory")
	}

	var challengeStartDateParsed time.Time
	if challengeStartDate != "" {
		var err error
		challengeStartDateParsed, err = time.Parse(time.DateOnly, challengeStartDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to parse challengeStartDate")
		}
	} else {
		return nil, fiber.NewError(fiber.StatusBadRequest, "ChallengeStartDate is mandatory")
	}

	completedTasks := []string{}
	if len(task.CompletedTasks) > 0 {
		completedTasks = task.CompletedTasks
	}

	id := primitive.NewObjectID()
	taskData := entity.TasksEntity{
		Id:                id,
		UserId:            userId,
		Level:             task.Level,
		Day:               task.Day,
		Status:            task.Status,
		ChalengeStartDate: challengeStartDateParsed,
		Date:              date,
		CompletedTasks:    completedTasks,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	_, err := db.GetCollection("task").InsertOne(ctx, taskData)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error inserting task data.")
	}

	return &id, nil
}

func GenerateJWTToken(user entity.CustomerEntity) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return "", fiber.NewError(fiber.StatusInternalServerError, "JWT secret key not found.")
	}

	claims := jtoken.MapClaims{
		"Id":    user.Id.Hex(),
		"email": user.Email,
		"role":  "customer",
		"exp":   time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
