package auth

import (
	"challenge/database"
	"challenge/graph/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ChangeStatus(ctx context.Context, db *database.DB, input model.ChangeStatusRequestInput) *model.ResponseModel {
	
	var (
		taskColl = db.GetCollection("task")
	)

	objID, err := primitive.ObjectIDFromHex(input.ChallengeID)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Invalid challengeId",
		}
	}

	filter := bson.M{"_id": objID}

	update := bson.M{
		"$set": bson.M{
			"status":    input.Status,
			"updatedAt": time.Now().UTC(),
		},
	}

	updateRes, err := taskColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return &model.ResponseModel{
			Status:  false,
			Message: "Failed to update challenge data in MongoDB: " + err.Error(),
		}
	}

	if updateRes.MatchedCount == 0 {
		return &model.ResponseModel{
			Status:  false,
			Message: "challenge not found",
		}
	}

	return &model.ResponseModel{
		Status:  true,
		Message: "Status Updated",
	}
}
