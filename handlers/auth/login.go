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
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
			return nil, gqlerror.Errorf("Invalid credentials")
		}
		return nil, gqlerror.Errorf("Error occurred while fetching user: " + err.Error())
	}

	sessionID := primitive.NewObjectID().Hex()
	if customer.SessionId != "" && customer.SessionId != sessionID {
		update := bson.M{
			"$set": bson.M{
				"sessionId": sessionID,
			},
		}

		_, err := customerColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update active device ID: " + err.Error())
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(strings.TrimSpace(input.Password)))
	if err != nil {
		return nil, gqlerror.Errorf("Invalid credentials")
	}

	var tasks []entity.TasksEntity
	taskFilter := bson.M{"userId": customer.Id, "cycleCount": customer.CycleCount}
	cursor, err := db.GetCollection("task").Find(ctx, taskFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching tasks: " + err.Error())
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var task entity.TasksEntity
		if err := cursor.Decode(&task); err != nil {
			return nil, gqlerror.Errorf("Error decoding task: " + err.Error())
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Error iterating through tasks: " + err.Error())
	}

	var challenges []*model.Challenge
	for _, task := range tasks {
		if len(task.CompletedTasks) > 0 && task.CompletedTasks[0] != "" {
			challenges = append(challenges, &model.Challenge{
				ID:             task.Id.Hex(),
				Level:          task.Level,
				Day:            task.Day,
				UserID:         task.UserId.Hex(),
				Date:           task.Date.Format(time.DateOnly),
				CompletedTasks: task.CompletedTasks,
				Status:         task.Status,
			})
		}
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return nil, gqlerror.Errorf("JWT secret key not found.")
	}

	claims := jtoken.MapClaims{
		"Id":        customer.Id,
		"email":     customer.Email,
		"role":      "customer",
		"sessionId": sessionID,
		"exp":       time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, gqlerror.Errorf("Failed to generate JWT token: " + err.Error())
	}

	return &model.LoginResponse{
		User: &model.User{
			ID:       customer.Id.Hex(),
			UserName: customer.UserName,
			Email:    customer.Email,
			Token:    signedToken,
		},
		ChallengeStartDate: tasks[0].ChalengeStartDate.Format(time.DateOnly),
		Challenges:         challenges,
	}, nil
}
