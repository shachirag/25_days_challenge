package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	jtoken "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)
func LoginCustomer(ctx context.Context, db *database.DB, input model.LoginRequestInput) (*model.LoginResponse, error) {
	customerColl := db.GetCollection("user")

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail}

	var customer entity.CustomerEntity
	err := customerColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error occurred while fetching user: "+err.Error())
	}

	// Check if user is already logged in from another device
	if customer.ActiveDeviceId != "" && customer.ActiveDeviceId != input.DeviceID {
		// Update the active device ID to the current login device ID
		update := bson.M{
			"$set": bson.M{
				"activeDeviceId": input.DeviceID,
			},
		}

		_, err := customerColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, update)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update active device ID: "+err.Error())
		}

		// Optionally: You can send a notification or perform any cleanup for the previous session here.
	}

	// Compare hashed password
	err = bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(strings.TrimSpace(input.Password)))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Fetch user's tasks
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

	// Prepare challenges
	challenges := make([]*model.Challenge, len(tasks))
	for i, task := range tasks {
		completedTasks := []string{}
		if len(task.CompletedTasks) > 0 {
			completedTasks = task.CompletedTasks
		}
		challenges[i] = &model.Challenge{
			ID:             task.Id.Hex(),
			Level:          task.Level,
			Day:            task.Day,
			UserID:         task.UserId.Hex(),
			Date:           task.Date.Format(time.DateOnly),
			CompletedTasks: completedTasks,
			Status:         task.Status,
		}
	}

	// Generate JWT token
	_secret := os.Getenv("JWT_SECRET_KEY")
	if _secret == "" {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "JWT_SECRET_KEY not configured")
	}

	month := (time.Hour * 24) * 30
	claims := jtoken.MapClaims{
		"Id":       customer.Id.Hex(),
		"email":    smallEmail,
		"role":     "customer",
		"deviceId": input.DeviceID, 
		"exp":      time.Now().Add(month * 6).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	_token, err := token.SignedString([]byte(_secret))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to generate JWT token: "+err.Error())
	}

	update := bson.M{
		"$set": bson.M{
			"activeDeviceId": input.DeviceID,
		},
	}

	_, err = customerColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, update)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update active device ID: "+err.Error())
	}

	return &model.LoginResponse{
		User: &model.User{
			ID:       customer.Id.Hex(),
			UserName: customer.UserName,
			Email:    customer.Email,
			Token:    _token,
		},
		ChallengeStartDate: tasks[0].ChalengeStartDate.Format(time.DateOnly), 
		Challenges:         challenges,
	}, nil
}
