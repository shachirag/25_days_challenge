package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ChangeStatus(ctx context.Context, db *database.DB, input model.ChangeStatusRequestInput) (*model.Challenge, error) {
	var (
		taskColl = db.GetCollection("task")
	)

	objID, err := primitive.ObjectIDFromHex(input.ChallengeID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID")
	}

	filter := bson.M{"_id": objID}

	var task entity.TasksEntity
	err = db.GetCollection("task").FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Task not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error occurred while fetching task: "+err.Error())
	}

	update := bson.M{
		"$set": bson.M{
			"status":    input.Status,
			"updatedAt": time.Now().UTC(),
		},
	}

	updateRes, err := taskColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update challenge data in MongoDB: "+err.Error())
	}

	if updateRes.MatchedCount == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "Challenge not found")
	}

	completedTasks := []string{}
	if len(task.CompletedTasks) > 0 {
		completedTasks = task.CompletedTasks
	}

	return &model.Challenge{
		ID:             task.Id.Hex(),
		Level:          task.Level,
		Day:            task.Day,
		Date:           task.Date.Format(time.DateOnly),
		Status:         task.Status,
		UserID:         task.UserId.Hex(),
		CompletedTasks: completedTasks,
	}, nil
}
