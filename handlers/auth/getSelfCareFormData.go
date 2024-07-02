package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSelfCareFormData(ctx context.Context, db *database.DB, input model.GetSelfCareFormRequestInput) (*model.SelfCareReponse, error) {

	user, err := utils.ExtractUserFromContext(ctx, db)
	if err != nil {
		return nil, err
	}

	deviceID, err := utils.ExtractDeviceIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if user.SessionId != deviceID {
		return nil, gqlerror.Errorf("Action not allowed. You are logged in from another device.")
	}

	var task entity.TasksEntity

	taskColl := db.GetCollection("task")

	filter := bson.M{
		"userId": user.Id,
		"day":    input.Day,
		"level":  input.Level,
	}

	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("task not found")
		}
		return nil, gqlerror.Errorf("failed to fetch task")
	}

	var OwnStatement string
	var thingsYouWillDoTodayToNutureYourself []string
	var thingsThatYouLoveAboutYourself []string
	var thingsTodayThatBringYouJoyAndFlow []string
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
