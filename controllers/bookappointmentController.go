package controller

import (
	"context"
	"fmt"
	"golang-hospital-management/database"
	"golang-hospital-management/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

var appointmentCollection *mongo.Collection = database.OpenCollection(database.Client, "appointment")

func GetAppoinments() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := appointmentCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing appointments"})
		}
		var allAppointment []bson.M
		if err = result.All(ctx, &allAppointment); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allAppointment)
	}
}

func GetAppoinment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		appointmentId := c.Param("appointment_id")
		var Appointment models.Appointment

		err := appointmentCollection.FindOne(ctx, bson.M{"order_id": appointmentId}).Decode(&Appointment)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the appointments"})
		}
		c.JSON(http.StatusOK, Appointment)
	}
}

func CreateAppointment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var doctor models.Doctor

		var appointment models.Appointment

		if err := c.BindJSON(&appointment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(appointment)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if appointment.Doctor_id != nil {
			err := appointmentCollection.FindOne(ctx, bson.M{"Doctor_id": appointment.Doctor_id}).Decode(&doctor)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Doctor not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}

		appointment.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		appointment.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		appointment.ID = primitive.NewObjectID()
		appointment.Appointment_id = appointment.ID.Hex()

		result, insertErr := appointmentCollection.InsertOne(ctx, appointment, nil)

		if insertErr != nil {
			msg := fmt.Sprintf("appointment was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func UpdateAppointment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var doctor models.Doctor
		var appointment models.Appointment

		var updateObj primitive.D

		appoinmentId := c.Param("appointment_id")
		if err := c.BindJSON(&appointment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if appointment.Doctor_id != nil {
			err := appointmentCollection.FindOne(ctx, bson.M{"appointment_id": appointment.Doctor_id}).Decode(&doctor)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Appointment was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"appointment", appointment.Doctor_id})
		}

		appointment.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", appointment.Updated_at})

		upsert := true

		filter := bson.M{"appointment_id": appoinmentId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := appointmentCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$st", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("appointment update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
