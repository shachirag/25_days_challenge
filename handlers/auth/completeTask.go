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

func CompleteTask(ctx context.Context, db *database.DB, input model.ChangeStatusRequestInput) (*model.Challenge, error) {

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	deviceID, err := utils.ExtractDeviceIDFromContext(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to extract device ID from context")
	}

	// fmt.Println(deviceID)
	// fmt.Println(user.SessionId)

	if user.SessionId != deviceID {
		return nil, gqlerror.Errorf("Action not allowed. You are logged in from another device.")
	}

	var task entity.TasksEntity
	taskColl := db.GetCollection("task")
	date, err := utils.ParseDate(input.Date)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid date format")
	}

	filter := bson.M{
		"userId": user.Id,
		"level":  input.Level,
		"day":    input.Day,
		"date":   date,
	}

	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			newTask := entity.TasksEntity{
				Id:             primitive.NewObjectID(),
				UserId:         user.Id,
				Level:          input.Level,
				Day:            input.Day,
				Date:           date,
				Status:         "incomplete",
				CompletedTasks: []string{input.CompletedTask},
				UpdatedAt:      time.Now().UTC(),
			}

			_, err = taskColl.InsertOne(ctx, newTask)
			if err != nil {
				return nil, gqlerror.Errorf("Failed to create new task document: " + err.Error())
			}

			return &model.Challenge{
				ID:             newTask.Id.Hex(),
				Level:          input.Level,
				Day:            input.Day,
				Date:           input.Date,
				Status:         newTask.Status,
				UserID:         user.Id.Hex(),
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
		update["$set"].(bson.M)["status"] = "incomplete"
	}

	_, err = taskColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update task data in MongoDB: " + err.Error())
	}

	if update["$set"] != nil {
		task.Status = update["$set"].(bson.M)["status"].(string)
	}

	return &model.Challenge{
		ID:             task.Id.Hex(),
		Level:          input.Level,
		Day:            input.Day,
		Date:           input.Date,
		Status:         task.Status,
		UserID:         user.Id.Hex(),
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
