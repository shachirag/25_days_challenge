package auth

import (
	"challenge/database"
	"challenge/graph/model"
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeleteAccount(ctx context.Context, db *database.DB, userId string) (*model.Response, error) {

	if userId == "" {
		return nil, gqlerror.Errorf("userId is mandatory")
	}

	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, gqlerror.Errorf("failed to convert object Id")
	}

	filter := bson.M{"_id": userObjId}

	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
		},
	}

	updateRes, err := db.GetCollection("user").UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, gqlerror.Errorf("failed to delete user account")
	}

	if updateRes.MatchedCount == 0 {
		return nil, gqlerror.Errorf("No User found")
	}

	return &model.Response{
		Message: "Account Deleted Successfully",
	}, nil
}
