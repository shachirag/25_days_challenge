package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SelfCareForm(ctx context.Context, db *database.DB, input model.SelfCareFormRequestInput) (*model.SelfCareReponse, error) {

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

	var (
		taskColl = db.GetCollection("task")
		task     entity.TasksEntity
	)

	objID, err := primitive.ObjectIDFromHex(input.ChallengeID)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid challenge ID")
	}

	filter := bson.M{
		"_id": objID,
	}
	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, gqlerror.Errorf("No task found")
		}
		return nil, gqlerror.Errorf("Internal server error while fetching the task: " + err.Error())
	}

	thingsToNurtureYourself := strings.Split(input.ThingsYouWillDoTodayToNutureYourself, ",")
	thingsYouLoveAboutYourself := strings.Split(input.ThingsThatYouLoveAboutYourself, ",")
	thingsThatBringJoyAndFlow := strings.Split(input.ThingsTodayThatBringYouJoyAndFlow, ",")

	update := bson.M{
		"$set": bson.M{
			"selfCareForm.ownStatement":                         input.OwnStatement,
			"selfCareForm.thingsThatYouLoveAboutYourself":       thingsYouLoveAboutYourself,
			"selfCareForm.thingsTodayThatBringYouJoyAndFlow":    thingsThatBringJoyAndFlow,
			"selfCareForm.thingsYouWillDoTodayToNutureYourself": thingsToNurtureYourself,
		},
	}

	_, err = taskColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update the form")
	}

	return &model.SelfCareReponse{
		ID:                                   task.Id.Hex(),
		OwnStatement:                         input.OwnStatement,
		ThingsYouWillDoTodayToNutureYourself: thingsToNurtureYourself,
		ThingsThatYouLoveAboutYourself:       thingsYouLoveAboutYourself,
		ThingsTodayThatBringYouJoyAndFlow:    thingsThatBringJoyAndFlow,
	}, nil
}
