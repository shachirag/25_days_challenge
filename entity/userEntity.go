package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerEntity struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	UserName      string             `json:"userName" bson:"userName"`
	Email         string             `json:"email" bson:"email"`
	Password      string             `json:"password" bson:"password"`
	SocialDetails SocialDetails      `json:"socialDetails" bson:"socialDetails"`
	CreatedAt     time.Time          `json:"cresatedAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type SocialDetails struct {
	AppleId  string `json:"appleId"`
	GoogleId string `json:"googleId"`
}
