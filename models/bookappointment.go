package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Appointment struct {
	ID         primitive.ObjectID `bson:"_id"`
	Appointment_Date time.Time          `json:"Appointment_date" validate:"required"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Appointment_id   string             `json:"Appointment_id"`
	Invoice_id   *string            `json:"Invoice_id" validate:"required"`
	Prescription_id   string            `json:"Prescription_id"`
	Doctor_id    *string             `json:"doctor_id"`
	
}