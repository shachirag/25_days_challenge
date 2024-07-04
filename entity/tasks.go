package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TasksEntity struct {
	Id                primitive.ObjectID `json:"id" bson:"_id"`
	UserId            primitive.ObjectID `json:"userId" bson:"userId"`
	Level             int                `json:"level" bson:"level"`
	Day               int                `json:"day" bson:"day"`
	Status            string             `json:"status" bson:"status"`
	CompletedTasks    []string           `json:"completedTasks" bson:"completedTasks"`
	ChalengeStartDate time.Time          `json:"challengeStartDate" bson:"challengeStartDate"`
	Date              time.Time          `json:"date" bson:"date"`
	CycleCount        int64              `json:"cycleCount" bson:"cycleCount"`
	SelfCareForm      *SelfCareForm      `json:"selfCareForm,omitempty" bson:"selfCareForm,omitempty"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type SelfCareForm struct {
	OwnStatement                         string   `json:"ownStatement" bson:"ownStatement"`
	ThingsYouWillDoTodayToNutureYourself []string `json:"thingsYouWillDoTodayToNutureYourself" bson:"thingsYouWillDoTodayToNutureYourself"`
	ThingsThatYouLoveAboutYourself       []string `json:"thingsThatYouLoveAboutYourself" bson:"thingsThatYouLoveAboutYourself"`
	ThingsTodayThatBringYouJoyAndFlow    []string `json:"thingsTodayThatBringYouJoyAndFlow" bson:"thingsTodayThatBringYouJoyAndFlow"`
}
