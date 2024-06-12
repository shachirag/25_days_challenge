package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TasksEntity struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id"`
	UserId             primitive.ObjectID `json:"userId" bson:"userId"`
	Level              int                `json:"level" bson:"level"`
	Day                int                `json:"day" bson:"day"`
	Status             string             `json:"status" bson:"status"`
	CompletedTasks     []string           `json:"completedTasks" bson:"completedTasks"`
	ChalengeStartDate time.Time          `json:"chalengeStartDate" bson:"chalengeStartDate"`
	Date               time.Time          `json:"date" bson:"date"`
	CreatedAt          time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
}
