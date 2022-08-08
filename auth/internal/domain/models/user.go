package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Login    string             `json:"login" bson:"login"`
	Password string             `bson:"password"`
}
