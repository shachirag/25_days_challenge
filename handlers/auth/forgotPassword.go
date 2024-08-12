package auth

import (
	"challenge/database"
	"challenge/graph/model"
	"context"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
)

func ForgotPassword(ctx context.Context, db *database.DB, input model.ForgotPasswordRequestInput) (*model.ForgotPasswordResponse, error) {
	var (
		userColl = db.GetCollection("user")
	)

	smallEmail := strings.ToLower(input.Email)

	filter := bson.M{"email": smallEmail, "isDeleted": false}

	count, err := userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Internal server error while counting the documents.")
	}

	if count == 0 {
		return &model.ForgotPasswordResponse{
			IsUserFound: false,
		}, gqlerror.Errorf("User not found")
	}

	return &model.ForgotPasswordResponse{
		IsUserFound: true,
	}, nil
}
