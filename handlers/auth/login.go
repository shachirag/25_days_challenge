package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
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
			return nil, gqlerror.Errorf("Invalid credentials")
		}
		return nil, gqlerror.Errorf("Error occurred while fetching user: " + err.Error())
	}

	if customer.SessionId != "" && customer.SessionId != input.DeviceID {
		update := bson.M{
			"$set": bson.M{
				"sessionId": input.DeviceID,
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
	taskFilter := bson.M{"userId": customer.Id}
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

	token, err := GenerateJWTToken(customer)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to generate JWT token: " + err.Error())
	}

	update := bson.M{
		"$set": bson.M{
			"sessionId": input.DeviceID,
		},
	}

	_, err = customerColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update active device ID: " + err.Error())
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
