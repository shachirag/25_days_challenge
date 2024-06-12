package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"os"
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

	var tasks []entity.TasksEntity
	taskFilter := bson.M{"userId": customer.Id}
	cursor, err := db.GetCollection("task").Find(ctx, taskFilter)
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "An error occurred while fetching tasks",
		}
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var task entity.TasksEntity
		if err := cursor.Decode(&task); err != nil {
			return &model.LoginPayload{
				Status:  false,
				Message: "Error decoding task",
			}
		}
		tasks = append(tasks, task)
	}

	challenges := []*model.Challenge{}
	for _, task := range tasks {
		challenge := &model.Challenge{
			Level:          task.Level,
			Day:            task.Day,
			Date:           task.Date.Format(time.DateOnly),
			CompletedTasks: task.CompletedTasks,
			Status:         task.Status,
		}
		challenges = append(challenges, challenge)
	}

	_secret := os.Getenv("JWT_SECRET_KEY")
	month := (time.Hour * 24) * 30
	claims := jtoken.MapClaims{
		"Id":    customer.Id,
		"email": smallEmail,
		"role":  "customer",
		"exp":   time.Now().Add(month * 6).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	_token, err := token.SignedString([]byte(_secret))
	if err != nil {
		return &model.LoginPayload{
			Status:  false,
			Message: "failed to get token",
		}
	}

	return &model.LoginPayload{
		Status:  true,
		Message: "Login successful.",
		Data: &model.LoginResponse{
			User: &model.User{
				ID:       customer.Id.Hex(),
				UserName: customer.UserName,
				Email:    customer.Email,
				Token:    _token,
			},
			ChallengeStartDate: tasks[0].ChalengeStartDate.Format(time.DateOnly),
			Challenges:         challenges,
		},
	}
}
