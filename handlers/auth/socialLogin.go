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

func SocialLoginCustomer(ctx context.Context, db *database.DB, input model.SocialLoginRequestInput) (*model.LoginResponse, error) {
	var (
		userColl = db.GetCollection("user")
		customer *entity.CustomerEntity
	)

	err := utils.ValidateSocialId(&input)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid social ID: "+err.Error())
	}

	filter := bson.M{}
	switch input.Type {
	case "apple":
		filter = bson.M{"appleId": input.SocialID}
	case "google":
		filter = bson.M{"googleId": input.SocialID}
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Unsupported social login type")
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
						return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to sign up: "+err.Error())
					}
				} else {
					return nil, fiber.NewError(fiber.StatusInternalServerError, "Error finding user by email: "+err.Error())
				}
			}
		} else {
			return nil, fiber.NewError(fiber.StatusNotFound, "User not found with provided social ID and no email to fallback")
		}
	}

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

	var tasks []entity.TasksEntity
	taskFilter := bson.M{"userId": customer.Id}
	cursor, err := db.GetCollection("task").Find(ctx, taskFilter)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error occurred while fetching tasks: "+err.Error())
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var task entity.TasksEntity
		if err := cursor.Decode(&task); err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Error decoding task: "+err.Error())
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error iterating through tasks: "+err.Error())
	}

	challenges := []*model.Challenge{}
	for _, task := range tasks {
		completedTasks := []string{}
		if len(task.CompletedTasks) > 0 {
			completedTasks = task.CompletedTasks
		}
		challenge := &model.Challenge{
			ID:             task.Id.Hex(),
			Level:          task.Level,
			Day:            task.Day,
			UserID:         task.UserId.Hex(),
			Date:           task.Date.Format(time.DateOnly),
			CompletedTasks: completedTasks,
			Status:         task.Status,
		}
		challenges = append(challenges, challenge)
	}

	token, err := generateJWTToken(customer, smallEmail)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to generate JWT token: "+err.Error())
	}

	return &model.LoginResponse{
		User: &model.User{
			ID:       customer.Id.Hex(),
			UserName: customer.UserName,
			Email:    customer.Email,
			Token:    token,
		},
		ChallengeStartDate: tasks[0].ChalengeStartDate.Format(time.DateOnly),
		Challenges:         challenges,
	}, nil
}

func socialSignup(ctx context.Context, db *database.DB, data *model.SocialLoginRequestInput) (*entity.CustomerEntity, error) {
	userColl := db.GetCollection("user")

	var smallEmail string
	if data.Email != nil {
		smallEmail = strings.ToLower(*data.Email)
	}

	filter := bson.M{"email": smallEmail}
	exists, err := userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error checking existing user: "+err.Error())
	}
	if exists > 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "User with this email already exists")
	}

	id := primitive.NewObjectID()
	customer := &entity.CustomerEntity{
		Id:        id,
		Email:     smallEmail,
		UserName:  data.Name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	switch data.Type {
	case "google":
		customer.SocialDetails.GoogleId = data.SocialID
	case "apple":
		customer.SocialDetails.AppleId = data.SocialID
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Unsupported social login type")
	}

	_, err = userColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to insert new customer: "+err.Error())
	}

	return customer, nil
}

func generateJWTToken(customer *entity.CustomerEntity, email string) (string, error) {
	_secret := os.Getenv("JWT_SECRET_KEY")

	month := (time.Hour * 24) * 30
	claims := jwt.MapClaims{
		"Id":       customer.Id.Hex(),
		"email":    email,
		"role":     "customer",
		"deviceId": customer.ActiveDeviceId,
		"exp":      time.Now().Add(month * 6).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(_secret))
}
