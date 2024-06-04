package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerEntity struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	UserName      string             `json:"firstName" bson:"firstName"`
	Email         string             `json:"email" bson:"email"`
	Password      string             `json:"password" bson:"password"`
	SocialDetails SocialDetails      `json:"socialDetails" bson:"socialDetails"`
	CreatedAt     time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	// Image     string             `json:"image" bson:"image"`
}

type SocialDetails struct {
	AppleId  string `json:"appleId"`
	GoogleId string `json:"googleId"`
}
