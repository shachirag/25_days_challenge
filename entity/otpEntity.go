package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OtpEntity struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Otp       string             `json:"otp" bson:"otp"`
	Email     string             `json:"email" bson:"email"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
