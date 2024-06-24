package auth

import (
	"challenge/database"
	"challenge/entity"
	"challenge/graph/model"
	"challenge/utils"
	"context"
	"os"
	"strings"
	"time"

	jtoken "github.com/golang-jwt/jwt/v4"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SocialLoginCustomer(ctx context.Context, db *database.DB, input model.SocialLoginRequestInput) (*model.LoginResponse, error) {
	var (
		userColl = db.GetCollection("user")
		customer *entity.CustomerEntity
	)

	err := utils.ValidateSocialId(&input)
	if err != nil {
		return nil, gqlerror.Errorf("Invalid social ID: " + err.Error())
	}

	filter := bson.M{}
	switch input.Type {
	case "apple":
		filter = bson.M{"socialDetails.appleId": input.SocialID}
	case "google":
		filter = bson.M{"socialDetails.googleId": input.SocialID}
	default:
		return nil, gqlerror.Errorf("Unsupported social login type")
	}

	var smallEmail string
	if input.Email != nil {
		smallEmail = strings.ToLower(*input.Email)
	}

	err = userColl.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments && input.Email != nil {
			filter = bson.M{"email": smallEmail}
			err = userColl.FindOne(ctx, filter).Decode(&customer)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					customer, err = socialSignup(ctx, db, &input)
					if err != nil {
						return nil, gqlerror.Errorf("Failed to sign up: " + err.Error())
					}

					if customer == nil {
						return nil, gqlerror.Errorf("Internal server error")
					}

					taskID, err := createInitialTask(ctx, db, customer.Id)
					if err != nil {
						return nil, gqlerror.Errorf("Failed to create initial task: " + err.Error())
					}

					challenges := []*model.Challenge{
						{
							ID:             taskID.Hex(),
							UserID:         customer.Id.Hex(),
							Level:          1,
							Day:            1,
							Date:           time.Now().Format(time.DateOnly),
							CompletedTasks: []string{},
							Status:         "incomplete",
						},
					}

					secret := os.Getenv("JWT_SECRET_KEY")
					if secret == "" {
						return nil, gqlerror.Errorf("JWT secret key not found.")
					}

					claims := jtoken.MapClaims{
						"Id":        customer.Id.Hex(),
						"email":     customer.Email,
						"role":      "customer",
						"sessionId": customer.SessionId,
						"exp":       time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
					}

					token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
					signedToken, err := token.SignedString([]byte(secret))
					if err != nil {
						return nil, gqlerror.Errorf("Failed to generate JWT token: " + err.Error())
					}

					return &model.LoginResponse{
						User: &model.User{
							ID:       customer.Id.Hex(),
							UserName: customer.UserName,
							Email:    customer.Email,
							Token:    signedToken,
						},
						ChallengeStartDate: time.Now().UTC().Format(time.DateOnly),
						Challenges:         challenges,
					}, nil
				} else {
					return nil, gqlerror.Errorf("Error finding user by email: " + err.Error())
				}
			}
		} else {
			return nil, gqlerror.Errorf("User not found with provided social ID and no email to fallback")
		}
	}

	sessionID := primitive.NewObjectID().Hex()
	update := bson.M{"sessionId": sessionID}

	if customer != nil {

		isUpdate := false
		if customer.Email == "" && smallEmail != "" {
			update["email"] = smallEmail
			isUpdate = true
		}
		if customer.SocialDetails.AppleId == "" && input.Type == "apple" {
			update["socialDetails.appleId"] = input.SocialID
			isUpdate = true
		}
		if customer.SocialDetails.GoogleId == "" && input.Type == "google" {
			update["socialDetails.googleId"] = input.SocialID
			isUpdate = true
		}

		if isUpdate {
			_, err = userColl.UpdateOne(ctx, bson.M{"_id": customer.Id}, bson.M{"$set": update})
			if err != nil {
				return nil, gqlerror.Errorf("Failed to update user details: " + err.Error())
			}
		}
	}

	var tasks []entity.TasksEntity
	taskFilter := bson.M{"userId": customer.Id}
	cursor, err := db.GetCollection("task").Find(ctx, taskFilter)
	if err != nil {
		return nil, gqlerror.Errorf("Error occurred while fetching tasks: " + err.Error())
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var task entity.TasksEntity
		if err := cursor.Decode(&task); err != nil {
			return nil, gqlerror.Errorf("Error decoding task: " + err.Error())
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		return nil, gqlerror.Errorf("Error iterating through tasks: " + err.Error())
	}

	var challenges []*model.Challenge
	for _, task := range tasks {
		if len(task.CompletedTasks) > 0 && task.CompletedTasks[0] != "" {
			challenges = append(challenges, &model.Challenge{
				ID:             task.Id.Hex(),
				Level:          task.Level,
				Day:            task.Day,
				UserID:         task.UserId.Hex(),
				Date:           task.Date.Format(time.DateOnly),
				CompletedTasks: task.CompletedTasks,
				Status:         task.Status,
			})
		}
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return nil, gqlerror.Errorf("JWT secret key not found.")
	}

	claims := jtoken.MapClaims{
		"Id":        customer.Id.Hex(),
		"email":     customer.Email,
		"role":      "customer",
		"sessionId": sessionID,
		"exp":       time.Now().Add(6 * 30 * 24 * time.Hour).Unix(),
	}

	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, gqlerror.Errorf("Failed to generate JWT token: " + err.Error())
	}
	return &model.LoginResponse{
		User: &model.User{
			ID:       customer.Id.Hex(),
			UserName: customer.UserName,
			Email:    customer.Email,
			Token:    signedToken,
		},
		ChallengeStartDate: tasks[0].ChalengeStartDate.Format(time.DateOnly),
		Challenges:         challenges,
	}, nil
}

func socialSignup(ctx context.Context, db *database.DB, data *model.SocialLoginRequestInput) (*entity.CustomerEntity, error) {
	userColl := db.GetCollection("user")

	var smallEmail string
	if data.Email != nil {
		smallEmail = strings.ToLower(*data.Email)
	}

	filter := bson.M{"email": smallEmail}
	exists, err := userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, gqlerror.Errorf("Error checking existing user: " + err.Error())
	}
	if exists > 0 {
		return nil, gqlerror.Errorf("User with this email already exists")
	}

	id := primitive.NewObjectID()

	sessionID := primitive.NewObjectID().Hex()
	customer := &entity.CustomerEntity{
		Id:        id,
		Email:     smallEmail,
		UserName:  data.Name,
		SessionId: sessionID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	switch data.Type {
	case "google":
		customer.SocialDetails.GoogleId = data.SocialID
	case "apple":
		customer.SocialDetails.AppleId = data.SocialID
	default:
		return nil, gqlerror.Errorf("Unsupported social login type")
	}

	_, err = userColl.InsertOne(ctx, customer)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to insert new customer: " + err.Error())
	}

	return customer, nil
}

func createInitialTask(ctx context.Context, db *database.DB, userId primitive.ObjectID) (*primitive.ObjectID, error) {
	taskColl := db.GetCollection("task")

	taskID := primitive.NewObjectID()
	initialTask := entity.TasksEntity{
		Id:                taskID,
		UserId:            userId,
		Level:             1,
		Day:               1,
		Date:              time.Now().UTC(),
		ChalengeStartDate: time.Now().UTC(),
		Status:            "incomplete",
		CompletedTasks:    []string{},
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	_, err := taskColl.InsertOne(ctx, initialTask)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to create initial task: " + err.Error())
	}

	return &taskID, nil
}
