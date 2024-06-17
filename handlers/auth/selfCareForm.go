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

func SelfCareForm(ctx context.Context, db *database.DB, deviceId string, input model.SelfCareFormRequestInput) (*model.SelfCareReponse, error) {

	var (
		taskColl = db.GetCollection("task")
		task     entity.TasksEntity
	)

	// deviceID, err := extractDeviceIDFromToken(ctx)
	// if err != nil {
	// 	return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	// }

	objID, err := primitive.ObjectIDFromHex(input.ChallengeID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid challenge ID")
	}

	filter := bson.M{
		"_id": objID,
	}
	err = taskColl.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fiber.NewError(fiber.StatusNotFound, "No task found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Internal server error while fetching the task: "+err.Error())
	}

	// user, err := getUserByToken(ctx, db)
	// if err != nil {
	// 	return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	// }

	// if user.ActiveDeviceId != deviceID {
	// 	return nil, fiber.NewError(fiber.StatusUnauthorized, "User is already logged in from another device")
	// }

	update := bson.M{
		"$set": bson.M{
			"selfCareForm.ownStatement":                         input.OwnStatement,
			"selfCareForm.thingsThatYouLoveAboutYourself":       input.ThingsThatYouLoveAboutYourself,
			"selfCareForm.thingsTodayThatBringYouJoyAndFlow":    input.ThingsTodayThatBringYouJoyAndFlow,
			"selfCareForm.thingsYouWillDoTodayToNutureYourself": input.ThingsYouWillDoTodayToNutureYourself,
		},
	}

	_, err = taskColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update the form")
	}

	return &model.SelfCareReponse{
		ID:                                   task.Id.Hex(),
		OwnStatement:                         input.OwnStatement,
		ThingsYouWillDoTodayToNutureYourself: input.ThingsYouWillDoTodayToNutureYourself,
		ThingsThatYouLoveAboutYourself:       input.ThingsThatYouLoveAboutYourself,
		ThingsTodayThatBringYouJoyAndFlow:    input.ThingsTodayThatBringYouJoyAndFlow,
	}, nil
}

// func getUserByToken(ctx context.Context, db *database.DB) (*entity.CustomerEntity, error) {
// 	token := ctx.Value("user").(*jtoken.Token)
// 	claims, ok := token.Claims.(jtoken.MapClaims)
// 	if !ok {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to parse token claims")
// 	}
// 	userIDHex, ok := claims["Id"].(string)
// 	if !ok {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "UserId not found in token claims")
// 	}

// 	userID, err := primitive.ObjectIDFromHex(userIDHex)
// 	if err != nil {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "Invalid user ID in token claims")
// 	}

// 	customerColl := db.GetCollection("user")
// 	var user entity.CustomerEntity
// 	err = customerColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
// 	if err != nil {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch user details: "+err.Error())
// 	}
// 	return &user, nil
// }
