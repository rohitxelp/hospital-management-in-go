
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Prescription struct {
	ID         primitive.ObjectID `bson:"_id"`
	Drugs       string             `json:"name" validate:"required"`
	Dosage      string             `json:"category" validate:"required"`
	Start_Date *time.Time         `json:"start_date"`
	End_Date   *time.Time         `json:"end_date"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Prescription_id    string             `json:"food_id"`
	Doctor_id    string             `json:"doctor_id"`

}