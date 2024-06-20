package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CompleteTask(ctx context.Context, db *database.DB, userID string, input model.ChangeStatusRequestInput) (*model.Challenge, error) {

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid user ID")
	}

	userFilter := bson.M{
		"_id": userObjID,
	}
	var user entity.CustomerEntity
	err = db.GetCollection("user").FindOne(ctx, userFilter).Decode(&user)
	if err != nil {
		return nil, gqlerror.Errorf("failed to fetch user")
	}

	deviceID, err := utils.ExtractDeviceIDFromContext(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to extract device ID from context")
	}

	if user.SessionId != deviceID {
		return nil, gqlerror.Errorf("Action not allowed. You are logged in from another device.")
	}

	var task entity.TasksEntity
	taskColl := db.GetCollection("task")
	date, err := utils.ParseDate(input.Date)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid format")
	}

	filter := bson.M{
		"userId": userObjID,
		"level":  input.Level,
		"day":    input.Day,
		"date":   date,
	}

	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			newTask := entity.TasksEntity{
				Id:             primitive.NewObjectID(),
				UserId:         userObjID,
				Level:          input.Level,
				Day:            input.Day,
				Date:           date,
				Status:         "ongoing",
				CompletedTasks: []string{input.CompletedTask},
				UpdatedAt:      time.Now().UTC(),
			}

			_, err := taskColl.InsertOne(ctx, newTask)
			if err != nil {
				return nil, gqlerror.Errorf("Failed to create new task document: " + err.Error())
			}

			return &model.Challenge{
				ID:             newTask.Id.Hex(),
				Level:          input.Level,
				Day:            input.Day,
				Date:           input.Date,
				Status:         newTask.Status,
				UserID:         userID,
				CompletedTasks: newTask.CompletedTasks,
			}, nil
		}
		return nil, gqlerror.Errorf("Error occurred while fetching task: " + err.Error())
	}

	update := bson.M{
		"$addToSet": bson.M{
			"completedTasks": input.CompletedTask,
		},
		"$set": bson.M{
			"updatedAt": time.Now().UTC(),
		},
	}

	if shouldMarkCompleted(input.Level, append(task.CompletedTasks, input.CompletedTask)) {
		update["$set"].(bson.M)["status"] = "completed"
	} else {
		update["$set"].(bson.M)["status"] = "ongoing"
	}

	updateRes, err := taskColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update task data in MongoDB: " + err.Error())
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("Task not found")
	}

	task.Status = update["$set"].(bson.M)["status"].(string)

	return &model.Challenge{
		ID:             task.Id.Hex(),
		Level:          input.Level,
		Day:            input.Day,
		Date:           input.Date,
		Status:         task.Status,
		UserID:         userID,
		CompletedTasks: task.CompletedTasks,
	}, nil
}

func shouldMarkCompleted(level int, completedTasks []string) bool {
	switch level {
	case 1:
		return len(completedTasks) >= 4
	case 2:
		return len(completedTasks) >= 6
	case 3:
		return len(completedTasks) >= 8
	default:
		return false
	}
}
