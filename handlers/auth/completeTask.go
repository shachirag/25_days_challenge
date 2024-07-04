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

	userObjIdID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid user ID")
	}

	var customer entity.CustomerEntity
	userFilter := bson.M{
		"_id": userObjIdID,
	}
	err = db.GetCollection("user").FindOne(ctx, userFilter).Decode(&customer)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching user: " + err.Error())
	}

	deviceID, err := utils.ExtractDeviceIDFromContext(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to extract device ID from context")
	}

	if customer.SessionId != deviceID {
		return nil, gqlerror.Errorf("Action not allowed. You are logged in from another device.")
	}

	var task entity.TasksEntity
	taskColl := db.GetCollection("task")
	filter := bson.M{
		"userId":     userObjIdID,
		"level":      input.Level,
		"day":        input.Day,
		"cycleCount": customer.CycleCount,
	}

	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			newTask := entity.TasksEntity{
				Id:             primitive.NewObjectID(),
				UserId:         customer.Id,
				Level:          input.Level,
				Day:            input.Day,
				CycleCount:     customer.CycleCount,
				Date:           time.Now().UTC(),
				Status:         "incomplete",
				CompletedTasks: []string{input.CompletedTask},
				UpdatedAt:      time.Now().UTC(),
				CreatedAt:      time.Now().UTC(),
			}
			_, err = taskColl.InsertOne(ctx, newTask)
			if err != nil {
				return nil, gqlerror.Errorf("Failed to create new task document: " + err.Error())
			}
			return &model.Challenge{
				ID:             newTask.Id.Hex(),
				Level:          input.Level,
				Day:            input.Day,
				Date:           time.Now().UTC().Format(time.DateOnly),
				Status:         newTask.Status,
				UserID:         userObjIdID.Hex(),
				CompletedTasks: newTask.CompletedTasks,
			}, nil
		}
		return nil, gqlerror.Errorf("Error occurred while fetching task: " + err.Error())
	}

	var updatedCompletedTasks []string
	if task.CompletedTasks != nil {
		updatedCompletedTasks = append(task.CompletedTasks, input.CompletedTask)
	} else {
		updatedCompletedTasks = []string{input.CompletedTask}
	}

	update := bson.M{
		"$addToSet": bson.M{
			"completedTasks": input.CompletedTask,
		},
		"$set": bson.M{
			"updatedAt": time.Now().UTC(),
		},
	}

	session, err := db.GetMongoClient().StartSession()
	if err != nil {
		return nil, gqlerror.Errorf("Failed to start session")
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		if shouldMarkCompleted(input.Level, updatedCompletedTasks) {
			if input.Level == 3 && len(updatedCompletedTasks) == 8 {
				newCycle := task.CycleCount + 1
				customerUpdate := bson.M{"$set": bson.M{"cycleCount": newCycle}}
				_, err = db.GetCollection("user").UpdateOne(ctx, userFilter, customerUpdate)
				if err != nil {
					return nil, gqlerror.Errorf("Failed to update customer cycle count: " + err.Error())
				}

				update["$addToSet"].(bson.M)["completedTasks"] = input.CompletedTask
				update["$set"].(bson.M)["status"] = "completed"
				_, err = taskColl.UpdateOne(ctx, filter, update)
				if err != nil {
					return nil, gqlerror.Errorf("Failed to update task data in MongoDB: " + err.Error())
				}

				task.Status = "completed"

				newTask := entity.TasksEntity{
					Id:                primitive.NewObjectID(),
					UserId:            customer.Id,
					Level:             1,
					Day:               1,
					Date:              time.Now().UTC(),
					ChalengeStartDate: time.Now().UTC(),
					CycleCount:        newCycle,
					Status:            "incomplete",
					CompletedTasks:    []string{},
					UpdatedAt:         time.Now().UTC(),
					CreatedAt:         time.Now().UTC(),
				}

				_, err = taskColl.InsertOne(ctx, newTask)
				if err != nil {
					return nil, gqlerror.Errorf("Failed to create new reset task document: " + err.Error())
				}

				completedTasks := []string{}
				if len(newTask.CompletedTasks) > 0 {
					completedTasks = newTask.CompletedTasks
				}

				return &model.Challenge{
					ID:             newTask.Id.Hex(),
					Level:          1,
					Day:            1,
					Date:           time.Now().UTC().Format(time.DateOnly),
					Status:         newTask.Status,
					UserID:         userObjIdID.Hex(),
					CompletedTasks: completedTasks,
				}, nil
			}

			update["$set"].(bson.M)["status"] = "completed"
		} else {
			update["$set"].(bson.M)["status"] = "incomplete"
		}

		_, err = taskColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, gqlerror.Errorf("Failed to update task data in MongoDB: " + err.Error())
		}

		task.Status = update["$set"].(bson.M)["status"].(string)
		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, gqlerror.Errorf("Transaction failed: %v", err)
	}

	return &model.Challenge{
		ID:             task.Id.Hex(),
		Level:          input.Level,
		Day:            input.Day,
		Date:           task.Date.Format(time.DateOnly),
		Status:         task.Status,
		UserID:         userObjIdID.Hex(),
		CompletedTasks: updatedCompletedTasks,
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
