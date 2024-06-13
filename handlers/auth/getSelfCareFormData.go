package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSelfCareFormData(ctx context.Context, db *database.DB, userId string, input model.GetSelfCareFormRequestInput) (*model.SelfCareReponse, error) {

	var task entity.TasksEntity

	taskColl := db.GetCollection("task")

	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID")
	}

	filter := bson.M{
		"userId": userObjID,
		"day":    input.Day,
		"level":  input.Level,
	}

	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "task not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch task")
	}

	var OwnStatement string
	var thingsYouWillDoTodayToNutureYourself string
	var thingsThatYouLoveAboutYourself string
	var thingsTodayThatBringYouJoyAndFlow string
	if task.SelfCareForm != nil {
		OwnStatement = task.SelfCareForm.OwnStatement
		thingsYouWillDoTodayToNutureYourself = task.SelfCareForm.ThingsYouWillDoTodayToNutureYourself
		thingsThatYouLoveAboutYourself = task.SelfCareForm.ThingsThatYouLoveAboutYourself
		thingsTodayThatBringYouJoyAndFlow = task.SelfCareForm.ThingsTodayThatBringYouJoyAndFlow
	}

	taskRes := &model.SelfCareReponse{
		ID:                                   task.Id.Hex(),
		OwnStatement:                         OwnStatement,
		ThingsYouWillDoTodayToNutureYourself: thingsYouWillDoTodayToNutureYourself,
		ThingsThatYouLoveAboutYourself:       thingsThatYouLoveAboutYourself,
		ThingsTodayThatBringYouJoyAndFlow:    thingsTodayThatBringYouJoyAndFlow,
	}

	return taskRes, nil
}
